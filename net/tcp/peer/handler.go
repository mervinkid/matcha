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

// ChannelHandler is the interface provide necessary methods for channel event
// handle which invoked by pipeline.
// Method:
//  ChannelActivate will be invoked while connection is ready.
//  ChannelInActivate will be invoked after connection closed.
//  ChannelRead will be invoked while a message is ready.
//  ChannelError will be invoked while some exception happened.
type ChannelHandler interface {
	ChannelActivate(channel Channel) error
	ChannelInactivate(channel Channel) error
	ChannelRead(channel Channel, in interface{}) error
	ChannelError(channel Channel, channelErr error)
}

// FunctionalChannelHandler is a public implementation of ChannelHandler interface which
// support functional definition for business logic.
type FunctionalChannelHandler struct {
	HandleActivate   func(channel Channel) error
	HandleInactivate func(channel Channel) error
	HandleRead       func(channel Channel, in interface{}) error
	HandleError      func(channel Channel, err error)
}

func (h *FunctionalChannelHandler) ChannelActivate(channel Channel) error {
	if h.HandleActivate != nil {
		return h.HandleActivate(channel)
	}
	return nil
}

func (h *FunctionalChannelHandler) ChannelInactivate(channel Channel) error {
	if h.HandleInactivate != nil {
		return h.HandleInactivate(channel)
	}
	return nil
}

func (h *FunctionalChannelHandler) ChannelRead(channel Channel, in interface{}) error {
	if h.HandleRead != nil {
		return h.HandleRead(channel, in)
	}
	return nil
}

func (h *FunctionalChannelHandler) ChannelError(channel Channel, channelErr error) {
	if h.HandleError != nil {
		h.HandleError(channel, channelErr)
	}
}
