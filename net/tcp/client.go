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

package tcp

import (
	"errors"
	"net"
	"sync"

	"github.com/mervinkid/allspark/logging"
	"github.com/mervinkid/allspark/misc"
	"github.com/mervinkid/allspark/net/tcp/config"
	"github.com/mervinkid/allspark/net/tcp/peer"
	"github.com/mervinkid/allspark/parallel"
)

// Errors
var ClientNotRunningError = errors.New("client is not running")

// Client is the interface that wraps the basic method to implement a tcp network client.
type Client interface {
	misc.Lifecycle
	misc.Sync
	peer.SendMessage
}

// PipelineServer is the default implementation of Client interface which using
// DuplexPipeline for connection handling.
type pipelineClient struct {
	Config config.ClientConfig

	// Initializer
	Initializer peer.PipelineInitializer

	pipeline   peer.Pipeline
	running    bool
	stateMutex sync.RWMutex
	waitGroup  sync.WaitGroup
}

// Start will start client and connect to remote.
func (c *pipelineClient) Start() error {

	// Mutex
	c.stateMutex.Lock()
	defer c.stateMutex.Unlock()

	if c.running == true {
		// Only work while client is not running.
		return nil
	}

	remoteAddr := new(net.TCPAddr)
	remoteAddr.IP = c.Config.IP
	remoteAddr.Port = c.Config.Port

	dialer := net.Dialer{}
	dialer.Timeout = c.Config.Timeout
	conn, err := dialer.Dial("tcp", remoteAddr.String())
	if err != nil {
		// Dial failure.
		return err
	}

	// Setup tcp props.
	config.TryApplyTCPConfig(&c.Config.TCPConfig, conn.(*net.TCPConn))

	// Init and start pipeline for connection.
	pipeline, err := peer.InitPipeline(conn, c.Initializer)
	if err != nil {
		return err
	}
	if err := pipeline.Start(); err != nil {
		return err
	}

	// Start a goroutine for pipeline state watching.
	c.startPipelineWatcher(pipeline)

	// Update state
	c.pipeline = pipeline
	c.running = true
	c.waitGroup.Add(1)

	return nil
}

func (c *pipelineClient) startPipelineWatcher(pipeline peer.Pipeline) {
	parallel.NewGoroutine(func() {
		logging.Trace("PipelineWatcher for remote %s start.\n", pipeline.Remote().String())
		pipeline.Sync()
		if misc.LifecycleCheckRun(c) {
			misc.LifecycleStop(c)
		}
		logging.Trace("PipelineWatcher for remote %s stop.\n", pipeline.Remote().String())
	}).Start()
}

// Stop will stop client and disconnect from remote.
func (c *pipelineClient) Stop() {

	// Mutex
	c.stateMutex.Lock()
	defer c.stateMutex.Unlock()

	if !c.running {
		// Only work while client is running.
		return
	}

	// Stop
	if misc.LifecycleCheckRun(c.pipeline) {
		misc.LifecycleStop(c.pipeline)
	}

	// Update state
	c.pipeline = nil
	c.running = false
	c.waitGroup.Done()
}

// IsRunning returns true if client is running.
func (c *pipelineClient) IsRunning() bool {
	c.stateMutex.RLock()
	defer c.stateMutex.RUnlock()
	return c.running
}

// Sync block invoker goroutine until client stop.
func (c *pipelineClient) Sync() {
	c.waitGroup.Wait()
}

// Send data synchronized.
func (c *pipelineClient) Send(data interface{}) error {

	c.stateMutex.RLock()
	defer c.stateMutex.RUnlock()

	if c.running && c.pipeline != nil && c.pipeline.GetChannel() != nil {
		channel := c.pipeline.GetChannel()
		return channel.Send(data)
	}

	return ClientNotRunningError
}

// Send data async, the callback method will be invoked after data has been handled.
func (c *pipelineClient) SendFuture(data interface{}, callback func(err error)) {

	c.stateMutex.RLock()
	defer c.stateMutex.RUnlock()

	if !c.running && callback != nil {
		callback(errors.New("client is not running"))
		return
	}

	c.pipeline.GetChannel().SendFuture(data, callback)
}

// NewPipelineClient create a new PipelineClient instance with specified configuration and initializer.
func NewPipelineClient(cfg config.ClientConfig, initializer peer.PipelineInitializer) Client {
	return &pipelineClient{
		Config:      cfg,
		Initializer: initializer,
		running:     false,
	}
}
