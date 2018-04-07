// The MIT License (MIT)
//
// Copyright (c) 2018 Mervin
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package registry

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/mervinkid/allspark/logging"
	"github.com/mervinkid/allspark/misc"
	"github.com/mervinkid/allspark/task"
	"io"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

const (
	redisElectionTtl   = 6000
	redisElectionDelay = 3 * time.Second
)

type redisRegistry struct {
	// Props
	config Config
	// Runtime
	role              Role
	redisConn         redis.Conn
	electionScheduler task.Scheduler
	// State
	running    bool
	stateMutex sync.RWMutex
	waitGroup  sync.WaitGroup
}

func (redisRegistry) Type() string {
	return "redis"
}

func (r *redisRegistry) Start() error {
	r.stateMutex.Lock()
	defer r.stateMutex.Unlock()
	if !r.running {
		electionScheduler := task.NewFixedDelayScheduler(r.electionTask, redisElectionDelay)
		if err := misc.LifecycleStart(electionScheduler); err != nil {
			r.redisConn.Close()
			return err
		}
		r.electionScheduler = electionScheduler
		r.running = true
		r.waitGroup.Add(1)
	}

	return nil
}

func (r *redisRegistry) Stop() {
	r.stateMutex.Lock()
	defer r.stateMutex.Unlock()
	if r.running {
		if misc.LifecycleCheckRun(r.electionScheduler) {
			misc.LifecycleStop(r.electionScheduler)
			r.electionScheduler = nil
		}
		if r.redisConn != nil {
			r.releaseRole()
			r.redisConn.Close()
			r.redisConn = nil
		}
		r.running = false
		r.waitGroup.Done()
	}
}

func (r *redisRegistry) IsRunning() bool {
	r.stateMutex.RLock()
	defer r.stateMutex.RUnlock()
	return r.running
}

func (r *redisRegistry) Sync() {
	r.waitGroup.Wait()
}

func (r *redisRegistry) checkNodeId() {
	if r.config.NodeId == "" {
		timestamp := time.Now().UnixNano()
		random := rand.New(rand.NewSource(timestamp)).Int63()
		src := strconv.FormatInt(timestamp, 10) + strconv.FormatInt(random, 10)
		hash := md5.New()
		hash.Write([]byte(src))
		hashCode := hex.EncodeToString(hash.Sum(nil))
		r.config.NodeId = r.config.AppId + "-" + hashCode
	}
}

func (r *redisRegistry) checkConn() error {
	if r.redisConn != nil {
		_, err := r.redisConn.Do("PING")
		if err == nil {
			return nil
		}
		if err != nil && err != io.EOF {
			return err
		}
		if err == io.EOF {
			r.redisConn.Close()
			r.redisConn = nil
		}
	}
	host := r.config.Url.Host
	port := r.config.Url.Port
	conn, err := redis.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return err
	}
	r.redisConn = conn
	return nil
}

func (r *redisRegistry) electionKey() string {
	return fmt.Sprintf("%s/election", r.config.AppId)
}

func (r *redisRegistry) electionTask() {
	// Init node id
	r.checkNodeId()
	if err := r.checkConn(); err != nil {
		r.changeRole(Slaver, "unknown")
		return
	}

	if r.role == Master {
		// Valid role
		nodeId, err := redis.String(r.redisConn.Do("GET", r.electionKey()))
		if err != nil {
			fmt.Printf("Try get value fail cause %s.", err.Error())
			logging.Error("Try get value fail cause %s.", err.Error())
			r.changeRole(Slaver, "unknown")
			return
		}
		if nodeId == r.config.NodeId {
			// Refresh data
			result, err := redis.Int(r.redisConn.Do("PEXPIRE", r.electionKey(), redisElectionTtl))
			if err != nil {
				fmt.Printf("Refresh lock expire fail cause %s.\n", err.Error())
				logging.Error("Refresh lock expire fail cause %s.", err.Error())
				r.changeRole(Slaver, "unknown")
				return
			}
			if result == 1 {
				r.changeRole(Master, r.config.NodeId)
				return
			} else {
				r.changeRole(Slaver, "unknown")
				return
			}
		} else {
			r.changeRole(Slaver, nodeId)
			return
		}

	} else {
		getLock, err := r.redisConn.Do("SET", r.electionKey(), r.config.NodeId, "NX", "PX", redisElectionTtl)
		if err != nil {
			fmt.Printf("Try get lock fail cause %s.\n", err.Error())
			logging.Error("Try get lock fail cause %s.", err.Error())
			r.changeRole(Slaver, "unknown")
			return
		}
		if getLock == "OK" {
			// Take lead
			r.changeRole(Master, r.config.NodeId)
			return
		} else {
			// Lose lead
			// Get current lead data
			nodeId, err := redis.String(r.redisConn.Do("GET", r.electionKey()))
			if err != nil {
				fmt.Printf("Try get value fail cause %s.", err.Error())
				logging.Error("Try get value fail cause %s.", err.Error())
				r.changeRole(Slaver, "unknown")
				return
			}
			r.changeRole(Slaver, nodeId)
			return
		}
	}

}

func (r *redisRegistry) changeRole(newRole Role, newMaster string) {
	if r.role != newRole {
		r.role = newRole
		if r.config.Election != nil {
			if newRole == Slaver {
				logging.Debug("Node %s is slaver.", r.config.NodeId)
				r.config.Election(MasterLose, newMaster)
			} else {
				logging.Debug("Node %s is master.", r.config.NodeId)
				r.config.Election(MasterTake, newMaster)
			}
		}
	}
}

func (r *redisRegistry) releaseRole() {
	if r.role == Master {
		nodeId, err := redis.String(r.redisConn.Do("GET ", r.electionKey()))
		if err != nil {
			return
		}
		if nodeId == r.config.NodeId {
			// Release
			r.redisConn.Do("DEL", r.electionKey())
		}
	}
	r.role = Slaver
}
