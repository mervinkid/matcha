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

package peer

import (
	"github.com/mervinkid/matcha/buffer"
	"github.com/mervinkid/matcha/misc"
	"github.com/mervinkid/matcha/net/tcp/codec"
	"github.com/mervinkid/matcha/parallel"

	"errors"
	"github.com/mervinkid/matcha/logging"
	"net"
	"sync"
)

// Chan buffer
const (
	dataChanSize = 10
	cmdChanSize  = 2
)

// State of pipeline
const (
	stateNew      = iota
	stateReady
	stateRunning
	stateShutdown
)

// Buffer size
const (
	readBufferSize = 1024
	byteBufferSize = 2 * readBufferSize
)

// Errors
var (
	NilInitializerError = errors.New("initializer is nil")
	NilConnError        = errors.New("conn is nil")
	NilDecoderError     = errors.New("decoder is nil")
	NilEncoderError     = errors.New("encoder is nil")
	NilHandlerError     = errors.New("handler is nil")
)

// Pipeline is the interface defined necessary methods which makes a pipeline of FrameDecoder,
// FrameEncoder, and ChannelHandler for inbound and outbound data processing.
//
// Model:
//  +---------------------------------------+
//  |         TCP Network Connection        |
//  +---------------------------------------+
//          ↑(write)               ↓(read)
//  +----------------+     +----------------+
//  |  FrameEncoder  |     |  FrameDecoder  |
//  +----------------+     +----------------+
//          ↑(outbound)            ↓(inbound)
//  +----------------+     +----------------+
//  |    Channel     |     | ChannelHandler |
//  +----------------+     +----------------+
//
// State:
//  +-----+           +---------+          +----------+
//  | NEW | → Start → | RUNNING | → Stop → | SHUTDOWN |
//  +-----+           +---------+          +----------+
//
// Notes:
// The implementations should be parallel safe and based on FSM.
type Pipeline interface {
	misc.Lifecycle
	misc.Sync
	SendMessage
	GetChannel() Channel
	Remote() net.Addr
}

// DuplexPipeline is a implementation of Pipeline based on FSM and provide full duplex and
// non blocking processing for inbound and outbound data. Each pipeline will create three
// goroutine for data processing after start.
//
// Model:
//  +----------------------------------------------+
//  |            TCP Network Connection            |
//  +----------------------------------------------+
//          ↑(write)                      ↓(read)
//  +----------------+            +----------------+
//  |  FrameEncoder  |            |  FrameDecoder  |
//  +----------------+            +----------------+
//          ↑(push)                       ↓(produce)
//  +----------------+            +----------------+
//  | OutboundWorker |            | InboundBuffer  |
//  +----------------+            +----------------+
//          ↑(consume)                    ↓(consume)
//  +----------------+            +----------------+
//  | OutboundBuffer |            | InboundWorker  |
//  +----------------+            +----------------+
//          ↑(produce)                     ↓(push)
//  +----------------+            +----------------+
//  |    Channel     | ← relate → | ChannelHandler |
//  +----------------+            +----------------+
//
// State:
//  +-----+          +-------+           +---------+          +----------+
//  | NEW | → Init → | READY | → Start → | RUNNING | → Stop → | SHUTDOWN |
//  +-----+          +-------+           +---------+          +----------+
//
// Notes:
// Stop the pipeline will also close the tcp connection which bind with pipeline.
type duplexPipeline struct {
	encoder codec.FrameEncoder
	decoder codec.FrameDecoder
	handler ChannelHandler

	// Props
	conn    net.Conn // Setup while construct.
	channel Channel  // Setup after init.

	// State
	state          uint8
	stateMutex     sync.RWMutex
	stateWaitGroup sync.WaitGroup

	// Data chan
	inboundDataC  chan interface{}
	outboundDataC chan OutboundEntity

	// Handler command chan
	inboundHandlerStopC  chan uint8
	outboundHandlerStopC chan uint8

	// Handler coroutine
	connReadHandler parallel.Goroutine
	inboundHandler  parallel.Goroutine
	outboundHandler parallel.Goroutine
}

// InitPipeline create and init pipeline with initializer.
func InitPipeline(conn net.Conn, initializer PipelineInitializer) (Pipeline, error) {

	// Check arguments
	if conn == nil {
		return nil, NilConnError
	}
	if initializer == nil {
		return nil, NilInitializerError
	}

	// Init encoder, decoder and handler
	decoder := initializer.InitDecoder()
	logging.Trace("Init decoder for %s.\n", conn.RemoteAddr())
	encoder := initializer.InitEncoder()
	logging.Trace("Init encoder for %s.\n", conn.RemoteAddr())
	handler := initializer.InitHandler()
	logging.Trace("Init handler for %s.\n", conn.RemoteAddr())

	// New pipeline
	pipeline := &duplexPipeline{
		conn:    conn,
		decoder: decoder,
		encoder: encoder,
		handler: handler,
	}

	// Init pipeline
	if err := pipeline.Init(); err != nil {
		return nil, err
	}

	return pipeline, nil
}

// GetChannel returns the channel which created and bind with pipeline.
func (cp *duplexPipeline) GetChannel() Channel {
	return cp.channel
}

// Remote returns the remote address of connection with bind with pipeline.
func (cp *duplexPipeline) Remote() net.Addr {
	if cp.conn != nil {
		return cp.conn.RemoteAddr()
	}
	return &UnknownAddr{}
}

// Start only work while pipeline is in READ state. It will start three goroutine worker for
// inbound and outbound data processing and change state from READ to RUNNING.
func (cp *duplexPipeline) Start() error {

	cp.stateMutex.Lock()
	defer cp.stateMutex.Unlock()

	if cp.state != stateReady {
		// Only work while pipeline is in READY state.
		return nil
	}

	// Start handlers
	cp.startConnReadHandler()
	cp.startInboundHandler()
	cp.startOutboundHandler()

	cp.state = stateRunning
	cp.stateWaitGroup.Add(1)

	return nil
}

func (cp *duplexPipeline) startConnReadHandler() {

	coroutine := parallel.NewGoroutine(cp.handleConnRead)
	coroutine.Start()
	cp.connReadHandler = coroutine
}

func (cp *duplexPipeline) handleConnRead() {

	logging.Trace("ConnReadHandler for remote %s start.\n", cp.conn.RemoteAddr().String())
	defer logging.Trace("ConnReadHandler for remote %s stop.\n", cp.conn.RemoteAddr().String())

	// Channel activate
	if err := cp.handler.ChannelActivate(cp.channel); err != nil {
		cp.handler.ChannelError(cp.channel, err)
	}

	// Init buffer
	readBuffer := make([]byte, readBufferSize)
	byteBuffer := buffer.NewElasticUnsafeByteBuf(byteBufferSize)

	// Read bytes from connection
	for {
		count, err := cp.conn.Read(readBuffer)
		if err != nil {
			parallel.NewGoroutine(cp.Stop).Start()
			// Channel inactivate
			if err := cp.handler.ChannelInactivate(cp.channel); err != nil {
				cp.handler.ChannelError(cp.channel, err)
			}
			return
		}

		logging.Trace("ConnReadHandler read %d bytes from remote %s.\n", count, cp.conn.RemoteAddr().String())

		byteBuffer.WriteBytes(readBuffer[:count])
		for {
			result, err := cp.decoder.Decode(byteBuffer)
			if err != nil {
				cp.handler.ChannelError(cp.channel, err)
			} else if result != nil {
				cp.inboundDataC <- result
			} else {
				break
			}
		}
		byteBuffer.Release()

	}
}

func (cp *duplexPipeline) startInboundHandler() {

	coroutine := parallel.NewGoroutine(cp.handleInbound)
	coroutine.Start()
	cp.inboundHandler = coroutine
}

func (cp *duplexPipeline) handleInbound() {

	logging.Trace("InboundHandler for remote %s start.\n", cp.conn.RemoteAddr().String())

	defer func() {
		logging.Trace("InboundHandler for remote %s stop.\n", cp.conn.RemoteAddr().String())
	}()

	for {
		select {
		case inboundData := <-cp.inboundDataC:
			if err := cp.handler.ChannelRead(cp.channel, inboundData); err != nil {
				cp.handler.ChannelError(cp.channel, err)
			}
			continue
		case <-cp.inboundHandlerStopC:
			return
		}
	}
}

func (cp *duplexPipeline) startOutboundHandler() {

	coroutine := parallel.NewGoroutine(cp.handleOutbound)
	coroutine.Start()
	cp.outboundHandler = coroutine

}

func (cp *duplexPipeline) handleOutbound() {

	logging.Trace("OutboundHandler for remote %s start.", cp.conn.RemoteAddr().String())

	defer func() {
		logging.Trace("OutboundHandler for remote %s stop.", cp.conn.RemoteAddr().String())
	}()

	for {
		select {
		case outboundData := <-cp.outboundDataC:
			data := outboundData.Data
			callback := outboundData.Callback
			// Encode
			encodeResult, encodeErr := cp.encoder.Encode(data)
			if encodeErr != nil {
				cp.handler.ChannelError(cp.channel, encodeErr)
				if callback != nil {
					// Invoke callback
					callback(encodeErr)
				}
				continue
			}
			// Write
			writeCount, writeErr := cp.conn.Write(encodeResult)
			if callback != nil {
				// Invoke callback
				callback(writeErr)
				if writeErr == nil {
					logging.Trace("OutboundHandler write %d bytes to remote %s.",
						writeCount, cp.conn.RemoteAddr().String())
				}
				continue
			}
		case <-cp.outboundHandlerStopC:
			return
		}
	}
}

// Init make pipeline init and change it's state from NEW to READY.
func (cp *duplexPipeline) Init() error {

	cp.stateMutex.Lock()
	defer cp.stateMutex.Unlock()

	if cp.state == stateNew {

		// Check conn, codec and handler
		if cp.conn == nil {
			return NilConnError
		}
		if cp.decoder == nil {
			return NilDecoderError
		}
		if cp.encoder == nil {
			return NilEncoderError
		}
		if cp.handler == nil {
			return NilHandlerError
		}

		// Init data chan.
		cp.inboundDataC = make(chan interface{}, dataChanSize)
		cp.outboundDataC = make(chan OutboundEntity, dataChanSize)

		// Init handler command chan.
		cp.inboundHandlerStopC = make(chan uint8, cmdChanSize)
		cp.outboundHandlerStopC = make(chan uint8, cmdChanSize)

		// Init network channel and make it bind with current pipeline.
		cp.channel = NewChannel(cp)

		cp.state = stateReady
	}

	return nil
}

// Stop will stop pipeline and close connection.
func (cp *duplexPipeline) Stop() {

	// Mutex
	cp.stateMutex.Lock()
	defer cp.stateMutex.Unlock()

	if cp.state != stateRunning {
		return
	}

	// Send  stop cmd to handlers
	close(cp.inboundHandlerStopC)
	close(cp.outboundHandlerStopC)
	// Await termination
	cp.inboundHandler.Join()
	cp.outboundHandler.Join()

	// Close reader and connection
	cp.conn.Close()
	cp.connReadHandler.Join()

	// Close data channels
	close(cp.inboundDataC)
	close(cp.outboundDataC)

	// Change state
	cp.state = stateShutdown
	cp.stateWaitGroup.Done()

	// Cleanup runtime objects.
	cp.connReadHandler = nil
	cp.inboundHandler = nil
	cp.outboundHandler = nil
}

// IsRunning check whether or not it is running
func (cp *duplexPipeline) IsRunning() bool {

	cp.stateMutex.RLock()
	defer cp.stateMutex.RUnlock()

	return cp.state == stateRunning
}

// Send will put message object into outbound data queue and wait until message
// have been handled by outbound handler if pipeline current running.
func (cp *duplexPipeline) Send(msg interface{}) error {

	sendResultChan := make(chan error, 1)

	cp.SendFuture(msg, func(err error) {
		sendResultChan <- err
		close(sendResultChan)
	})

	return <-sendResultChan
}

// SendFuture put message object into outbound data queue and register callback
// function if pipeline current running. The callback function will be invoked
// by outbound handler after data processed.
func (cp *duplexPipeline) SendFuture(msg interface{}, callback func(err error)) {

	if msg == nil {
		return
	}

	cp.stateMutex.RLock()
	defer cp.stateMutex.RUnlock()

	if cp.state != stateRunning {
		if callback != nil {
			callback(errors.New("pipeline closed"))
		}
	}

	if cp.outboundDataC != nil {
		cp.outboundDataC <- OutboundEntity{
			Data:     msg,
			Callback: callback,
		}
	}
}

// Sync block invoker goroutine until pipeline stop.
func (cp *duplexPipeline) Sync() {
	cp.stateWaitGroup.Wait()
}
