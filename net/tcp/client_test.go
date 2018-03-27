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

package tcp_test

import (
	"fmt"
	"github.com/mervinkid/allspark/net/tcp"
	"github.com/mervinkid/allspark/net/tcp/codec"
	"github.com/mervinkid/allspark/net/tcp/config"
	"github.com/mervinkid/allspark/net/tcp/peer"
	"github.com/mervinkid/allspark/parallel"
	"log"
	"net"
	"testing"
	"time"
)

// Message definitions
type _tCommand struct {
	Id   int
	Name string
}

func (t *_tCommand) TypeCode() uint16 {
	return 1
}

func (t *_tCommand) String() string {
	return fmt.Sprintf("_tCommand{Id:%d}", t.Id)
}

type _tAck struct {
	Id int
}

func (t *_tAck) TypeCode() uint16 {
	return 2
}

func (t *_tAck) String() string {
	return fmt.Sprintf("_tAck{Id:%d}", t.Id)
}

func TestClient(t *testing.T) {

	// Client config
	clientConfig := config.ClientConfig{}
	clientConfig.KeepAlive = false
	clientConfig.IP = net.ParseIP("127.0.0.1")
	clientConfig.Port = 9090

	// Init client
	client := tcp.NewPipelineClient(clientConfig, initInitializer())
	client.Start()

	sender := parallel.NewGoroutine(func() {
		for i := 0; i < 10; i++ {
			msg := new(_tCommand)
			msg.Id = 12345 + i
			msg.Name = fmt.Sprint("TestCommand-", i)
			client.Send(msg)
			time.Sleep(1 * time.Second)
		}
	})
	sender.Start()
	sender.Sync()

	client.Stop()
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
		return new(_tCommand)
	})
	// Register _tAck
	apolloConfig.RegisterEntity(func() codec.ApolloEntity {
		return new(_tAck)
	})
	return apolloConfig
}

func initHandler() peer.ChannelHandler {
	handler := peer.FunctionalChannelHandler{}

	handler.HandleActivate = func(channel peer.Channel) error {
		log.Println(">>> Remote ", channel.Remote().String(), " activate.")
		return nil
	}

	handler.HandleRead = func(channel peer.Channel, in interface{}) error {
		log.Println(">>> Remote ", channel.Remote().String(), ": ", in)
		switch msg := in.(type) {
		case *_tCommand:
			mId := msg.Id
			ack := &_tAck{Id: mId}
			channel.Send(ack)
			log.Println(">>> Send ", ack)
		}
		return nil
	}

	handler.HandleInactivate = func(channel peer.Channel) error {
		log.Println(">>> Remote ", channel.Remote().String(), " inactivate.")
		return nil
	}

	handler.HandleError = func(channel peer.Channel, err error) {
		log.Panicln(">>> Remote ", channel.Remote().String(), " error: ", err)
	}
	return &handler
}
