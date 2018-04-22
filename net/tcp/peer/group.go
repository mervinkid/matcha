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
	"sync"

	"github.com/mervinkid/matcha/misc"
)

// ChannelGroup is a interface wraps methods for channel management which provide
// batch close support for channels.
type ChannelGroup interface {
	Add(channel Channel)
	Remove(channel Channel)
	CloseAll()
}

// HashSafeChannelGroup is a parallel safe implementation of ChannelGroup interface
// which based on hash-table.
type hashSafeChannelGroup struct {
	channelMap sync.Map
}

// Add will add a specified channel to channel group.
func (cg *hashSafeChannelGroup) Add(channel Channel) {
	if channel != nil {
		cg.channelMap.Store(channel, uint8(0))
	}
}

// Remove will remove specified channel from channel group.
func (cg *hashSafeChannelGroup) Remove(channel Channel) {
	if channel != nil {
		cg.channelMap.Delete(channel)
	}
}

// CloseAll will close all channel which management by channel group and remove
// all channel from channel group.
func (cg *hashSafeChannelGroup) CloseAll() {
	cg.channelMap.Range(func(key, value interface{}) bool {
		if channel, ok := key.(Channel); ok {
			misc.TryClose(channel)
		}
		cg.channelMap.Delete(key)
		return true
	})
}

// NewHashSafeChannelGroup create a instance of ChannelGroup based on hash-table.
func NewHashSafeChannelGroup() ChannelGroup {
	return &hashSafeChannelGroup{}
}
