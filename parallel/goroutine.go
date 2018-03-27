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

package parallel

import (
	"errors"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

// State constants
const (
	stateNew     = iota
	stateRunning
	stateFinish
)

// Errors
var IllegalStackFragmentError = errors.New("illegal stack fragment")

// Goroutine is the interface made definition of coroutine.
type Goroutine interface {
	Start()
	Join()
	IsAlive() bool
	GetId() uint64
}

type StatementGoroutine struct {
	statement      func()
	state          uint8
	stateMutex     sync.RWMutex
	stateWaitGroup sync.WaitGroup
	gId            uint64
}

// Start will start coroutine.
func (c *StatementGoroutine) Start() {
	c.stateMutex.Lock()
	defer c.stateMutex.Unlock()

	if c.state != stateNew {
		// Only work if coroutine is in NEW state.
		return
	}

	c.stateWaitGroup.Add(1)
	c.run()

	c.state = stateRunning
}

// IsAlive returns true is coroutine is in running state.
func (c *StatementGoroutine) IsAlive() bool {
	c.stateMutex.RLock()
	defer c.stateMutex.RUnlock()
	return c.state == stateRunning
}

func (c *StatementGoroutine) GetId() uint64 {
	return c.gId
}

// Sync block invoker goroutine until coroutine finish.
func (c *StatementGoroutine) Join() {
	c.stateWaitGroup.Wait()
}

// Run will execute statement. This method can be override with own logic when writing custom implementation.
func (c *StatementGoroutine) Run() {
	if c.statement != nil {
		c.statement()
	}
}

// run check and execute statement in goroutine and watch the state of the goroutine.
func (c *StatementGoroutine) run() {

	go func() {
		// Try get goroutine on start
		c.gId, _ = GetGoroutineId()
		// Execute statement
		c.Run()
		// Change state to FINISH
		c.stateMutex.Lock()
		c.state = stateFinish
		c.stateMutex.Unlock()
		// Release sync wait.
		c.stateWaitGroup.Done()
		// Cleanup goroutine context
		globalGoroutineLocalRepo.cleanupContext(c.gId)
	}()
}

// Create a Goroutine instance with statement function.
func NewGoroutine(statement func()) Goroutine {
	return &StatementGoroutine{statement: statement}
}

// GetGoroutineId returns id of invoker goroutine.
func GetGoroutineId() (uint64, error) {

	// Init read buffer and read stack information fragment.
	readBuffer := make([]byte, 64)
	count := runtime.Stack(readBuffer, false)
	stackFragment := string(readBuffer[:count])

	// Split fragment string by space.
	stackFragmentParts := strings.Split(stackFragment, " ")

	// Check split result
	if len(stackFragmentParts) < 2 {
		return 0, IllegalStackFragmentError
	}

	// Convert string value to int.
	goroutineId, err := strconv.ParseUint(stackFragmentParts[1], 10, 64)

	if err != nil {
		return 0, err
	}

	return goroutineId, nil
}

// GoroutineLocalRepository implement a parallel-safe repository for goroutine context with RWMutex.
//  +-----------------------------+
//  |  GID  |      Context        |
//  +-------+---------------------+
//  |       | +-----------------+ |
//  |       | |  KeyA  | ValueA | |
//  |       | +--------+--------+ |
//  |   1   | |  KeyB  | ValueB | |
//  |       | +--------+--------+ |
//  |       | |   ...  |   ...  | |
//  |       | +-----------------+ |
//  +-----------------------------+
//  |       | +-----------------+ |
//  |       | |  KeyA  | ValueA | |
//  |       | +--------+--------+ |
//  |   2   | |  KeyB  | ValueB | |
//  |       | +--------+--------+ |
//  |       | |   ...  |   ...  | |
//  |       | +-----------------+ |
//  +-----------------------------+
//  |  ...  |         ...         |
//  +-----------------------------+
type goroutineLocalRepo struct {
	dataMap map[uint64]map[interface{}]interface{}
}

func (r *goroutineLocalRepo) getGoroutineLocal(goroutineId uint64, key interface{}) interface{} {
	entity := r.dataMap[goroutineId]
	if entity == nil {
		return nil
	}
	return entity[key]
}

func (r *goroutineLocalRepo) setGoroutineLocal(goroutineId uint64, key interface{}, value interface{}) {
	entity := r.dataMap[goroutineId]
	if entity == nil {
		entity = make(map[interface{}]interface{})
		r.dataMap[goroutineId] = entity
	}
	entity[key] = value
}

func (r *goroutineLocalRepo) cleanupContext(goroutineId uint64) {
	delete(r.dataMap, goroutineId)
}

var globalGoroutineLocalRepo = &goroutineLocalRepo{dataMap: make(map[uint64]map[interface{}]interface{})}

// SetGoroutineContext set data to goroutine local .
func SetGoroutineLocal(key, value interface{}) {

	if key == nil {
		return
	}
	// Get goroutine id and try to set value.
	gId, err := GetGoroutineId()
	if err == nil {
		globalGoroutineLocalRepo.setGoroutineLocal(gId, key, value)
	}
}

// GetGoroutineLocal get local context data of invoker goroutine.
func GetGoroutineLocal(key interface{}) interface{} {

	if key == nil {
		return nil
	}

	gId, err := GetGoroutineId()
	if err == nil {
		return globalGoroutineLocalRepo.getGoroutineLocal(gId, key)
	}
	return nil
}
