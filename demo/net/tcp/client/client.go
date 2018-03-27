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

package main

import (
	"fmt"
	"math/rand"
	"net"
	"sync"
	"time"

	"flag"
	"github.com/mervinkid/allspark/logging"
	"github.com/mervinkid/allspark/net/tcp"
	"github.com/mervinkid/allspark/net/tcp/codec"
	"github.com/mervinkid/allspark/net/tcp/config"
	"github.com/mervinkid/allspark/net/tcp/peer"
	"github.com/mervinkid/allspark/parallel"
	"github.com/mervinkid/allspark/task"
	"os"
	"runtime"
	"strconv"
	"strings"
)

// Message definitions
type tCommand struct {
	Id   int
	Name string
}

func (t *tCommand) TypeCode() uint16 {
	return 1
}

func (t *tCommand) String() string {
	return fmt.Sprintf("_tCommand{Id:%d}", t.Id)
}

type tAck struct {
	Id int
}

func (t *tAck) TypeCode() uint16 {
	return 2
}

func (t *tAck) String() string {
	return fmt.Sprintf("_tAck{Id:%d}", t.Id)
}

var ackManager = peer.NewAckManager()

func main() {

	// Parse command line args
	address := flag.String("h", "localhost:9090", "host to connect")
	parallelism := flag.Int("p", 1, "parallelism")
	cpu := flag.Int("c", 0, "cpu")
	debug := flag.Bool("d", false, "debug")
	help := flag.Bool("help", false, "show usage")
	flag.Parse()
	if *help {
		flag.Usage()
		os.Exit(0)
	}

	if *cpu > 0 {
		runtime.GOMAXPROCS(*cpu)
	}
	if *debug {
		logging.SetLogLevel(logging.LDebug)
	} else {
		logging.SetLogLevel(logging.LInfo)
	}

	hostParts := strings.Split(*address, ":")

	// Client config
	clientConfig := config.ClientConfig{}
	clientConfig.KeepAlive = false
	clientConfig.IP = net.ParseIP(hostParts[0])
	clientConfig.Port, _ = strconv.Atoi(hostParts[1])

	sendSuccess := 0
	sendFailure := 0
	sendCounterMutex := new(sync.Mutex)

	monitor := parallel.NewGoroutine(func() {
		for {
			sendCounterMutex.Lock()
			logging.Info("Send success: %d", sendSuccess)
			logging.Info("Send failure: %d", sendFailure)
			sendSuccess = 0
			sendFailure = 0
			sendCounterMutex.Unlock()
			time.Sleep(1 * time.Second)
		}
	})
	monitor.Start()

	clients := make([]tcp.Client, *parallelism)

	for i := 0; i < *parallelism; i++ {
		// Init client
		client := tcp.NewPipelineClient(clientConfig, initInitializer())
		if err := client.Start(); err != nil {
			logging.Error("Can not start client cause %s.", err.Error())
			os.Exit(0)
		}

		scheduleSender := task.NewFixedRateScheduler(func() {
			time.Sleep(200 * time.Millisecond)
			msg := new(tCommand)
			msg.Id = rand.Int()
			msg.Name = fmt.Sprint("TestCommand-", msg.Id)
			ackManager.InitAck(msg.Id)
			err := client.Send(msg)
			sendCounterMutex.Lock()
			if err != nil {
				sendFailure += 1
				sendCounterMutex.Unlock()
				return
			}
			sendCounterMutex.Unlock()
			_, err = ackManager.WaitAck(msg.Id, time.Second*5)
			sendCounterMutex.Lock()
			if err != nil {
				sendFailure += 1
				sendCounterMutex.Unlock()
				return
			}
			sendSuccess += 1
			sendCounterMutex.Unlock()
		}, 2*time.Second)
		scheduleSender.Start()

		clients[i] = client

	}

	time.Sleep(30 * time.Second)
	for _, client := range clients {
		client.Stop()
	}
	for _, client := range clients {
		client.Sync()
	}
}

func initInitializer() peer.PipelineInitializer {

	apolloConfig := initApolloConfig()
	initializer := peer.FunctionalPipelineInitializer{}

	// Setup decoder init function
	initializer.DecoderInit = func() codec.FrameDecoder {
		return codec.NewApolloFrameDecoder(apolloConfig)
	}

	// Setup encoder init function
	initializer.EncoderInit = func() codec.FrameEncoder {
		return codec.NewApolloFrameEncoder(apolloConfig)
	}

	// Setup handler init function
	initializer.HandlerInit = func() peer.ChannelHandler {
		return initHandler()
	}

	return &initializer
}

func initApolloConfig() codec.ApolloConfig {
	apolloConfig := codec.ApolloConfig{}
	// Register _tCommand
	apolloConfig.RegisterEntity(func() codec.ApolloEntity {
		return new(tCommand)
	})
	// Register _tAck
	apolloConfig.RegisterEntity(func() codec.ApolloEntity {
		return new(tAck)
	})
	return apolloConfig
}

func initHandler() peer.ChannelHandler {
	handler := peer.FunctionalChannelHandler{}

	handler.HandleActivate = func(channel peer.Channel) error {
		logging.Debug(">>> Remote %s activate.", channel.Remote().String())
		return nil
	}

	handler.HandleRead = func(channel peer.Channel, in interface{}) error {
		logging.Debug(">>> Remote %s: %v.", channel.Remote().String(), in)
		switch msg := in.(type) {
		case *tCommand:
			mId := msg.Id
			ack := &tAck{Id: mId}
			channel.Send(ack)
			logging.Debug(">>> Send %v.", ack)
			break
		case *tAck:
			ackManager.CommitAck(msg.Id, nil)
			break
		}
		return nil
	}

	handler.HandleInactivate = func(channel peer.Channel) error {
		logging.Debug(">>> Remote %s inactivate.", channel.Remote().String())
		return nil
	}

	handler.HandleError = func(channel peer.Channel, err error) {
		logging.Warn(">>> Remote %s error: %s", channel.Remote().String(), err)
	}
	return &handler
}
