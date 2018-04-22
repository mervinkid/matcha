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
	"errors"
	"net"

	"github.com/mervinkid/matcha/misc"
)

const (
	unknownString = "unknown"
)

var (
	ErrInvalidChannel = errors.New("invalid channel")
)

type SendMessage interface {
	Send(data interface{}) error
	SendFuture(data interface{}, callback func(err error))
}

type Channel interface {
	SendMessage
	misc.Close
	Remote() net.Addr
	IsConnected() bool
	GetContext(key string) interface{}
	AddContext(key string, val interface{})
	DelContext(key string)
}

// PipelineChannel is a implementation of Channel interface created and bind with pipeline.
// It contact with pipeline by using a data chan.
// +------------+          +------------+
// |  Pipeline  | ← chan ← |  Channel   |
// +------------+          +------------+
type pipelineChannel struct {
	pipeline   Pipeline
	contextMap map[string]interface{}
}

// Remote returns remote address.
func (c *pipelineChannel) Remote() net.Addr {
	if c.pipeline != nil {
		return c.pipeline.Remote()
	}
	return &UnknownAddr{}
}

func (c *pipelineChannel) Send(data interface{}) error {

	if c.pipeline != nil && c.pipeline.IsRunning() {
		return c.pipeline.Send(data)
	}
	return ErrInvalidChannel
}

// SendFuture send data async and the callback method will be invoked after data have been write to connection.
func (c *pipelineChannel) SendFuture(data interface{}, callback func(err error)) {

	if c.pipeline != nil && c.pipeline.IsRunning() {
		c.pipeline.SendFuture(data, callback)
		return
	}

	if callback != nil {
		callback(ErrInvalidChannel)
	}
}

// Close will try close the network connection which related with current channel.
func (c *pipelineChannel) Close() {
	if c.pipeline != nil {
		c.pipeline.Stop()
	}
}

// IsConnected returns true if connection is valid.
func (c *pipelineChannel) IsConnected() bool {
	return c.pipeline != nil && c.pipeline.IsRunning()
}

// GetContext get context data with specified key.
func (c *pipelineChannel) GetContext(key string) interface{} {
	if c.contextMap != nil {
		return c.contextMap[key]
	}
	return nil
}

// AddContext add context data with specified key.
func (c *pipelineChannel) AddContext(key string, val interface{}) {
	if c.contextMap != nil {
		c.contextMap[key] = val
	}
}

// DelContext remove context data with specified key.
func (c *pipelineChannel) DelContext(key string) {
	if c.contextMap != nil {
		delete(c.contextMap, key)
	}
}

func NewChannel(pipeline Pipeline) Channel {

	return &pipelineChannel{
		pipeline:   pipeline,
		contextMap: make(map[string]interface{}),
	}
}

type UnknownAddr struct {
}

func (ua *UnknownAddr) String() string {
	return unknownString
}

func (ua *UnknownAddr) Network() string {
	return unknownString
}

type OutboundEntity struct {
	Data     interface{}
	Callback func(err error)
}
