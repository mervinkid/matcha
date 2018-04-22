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
	"flag"
	"fmt"
	"github.com/mervinkid/matcha/logging"
	"github.com/mervinkid/matcha/net/tcp"
	"github.com/mervinkid/matcha/net/tcp/codec"
	"github.com/mervinkid/matcha/net/tcp/config"
	"github.com/mervinkid/matcha/net/tcp/peer"
	"github.com/mervinkid/matcha/task"
	"os"
	"runtime"
	"time"
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

func main() {

	// Parse command line argument
	port := flag.Int("p", 9090, "port to listen")
	cpu := flag.Int("c", 0, "cpu")
	debug := flag.Bool("d", false, "debug")
	help := flag.Bool("help", false, "show usage")
	flag.Parse()
	if *help {
		flag.Parse()
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

	// Monitor
	scheduler := task.NewFixedRateScheduler(func() {
		memState := new(runtime.MemStats)
		var lastNumGC uint32
		runtime.ReadMemStats(memState)
		allocKB := memState.Alloc / 1024
		numGC := memState.NumGC - lastNumGC
		lastNumGC = memState.NumGC
		logging.Info("Monitor Alloc %dKB, NumGC %d.", allocKB, numGC)
	}, 2*time.Second)
	scheduler.Start()

	serverConfig := config.ServerConfig{}
	serverConfig.AcceptorSize = 2
	serverConfig.Port = *port

	server := tcp.NewPipelineServer(serverConfig, initInitializer())
	if err := server.Start(); err != nil {
		logging.Error("Cannot start server cause %s.", err.Error())
		os.Exit(0)
	}
	server.Sync()
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
		}
		return nil
	}

	handler.HandleInactivate = func(channel peer.Channel) error {
		logging.Debug(">>> Remote %s inactivate.", channel.Remote().String())
		return nil
	}

	handler.HandleError = func(channel peer.Channel, err error) {
		logging.Warn(">>> Remote %s error: %s.", channel.Remote().String(), err.Error())
	}
	return &handler
}
