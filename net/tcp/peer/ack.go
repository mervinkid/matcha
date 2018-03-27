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
	"sync"
	"time"
)

var AckTimeoutError = errors.New("ack timeout")

// AckManager is the interface wraps methods for acknowledgement management.
// Methods:
//  InitAck init and register a ack transaction to manager.
//  WaitAck will block invoker goroutine until specified ack transaction commit or timeout.
//  CommitAck commit specified ack transaction.
type AckManager interface {
	InitAck(key interface{})
	WaitAck(key interface{}, timeout time.Duration) (data interface{}, err error)
	CommitAck(key interface{}, data interface{})
}

// SafeAckManager is a parallel-safe implementation of AckManager interface.
type SafeAckManager struct {
	ackRespChanMap sync.Map
}

type ackRespEntity struct {
	data interface{}
	err  error
}

type ackRespChan chan ackRespEntity

// InitAck init and register a ack transaction to manager.
func (m *SafeAckManager) InitAck(key interface{}) {

	if key == nil {
		return
	}

	if _, ok := m.ackRespChanMap.Load(key); !ok {
		m.ackRespChanMap.Store(key, make(ackRespChan, 2))
	}
}

// WaitAck will block invoker goroutine until specified ack transaction commit or timeout.
func (m *SafeAckManager) WaitAck(key interface{}, timeout time.Duration) (interface{}, error) {

	if key == nil {
		return nil, nil
	}

	if value, ok := m.ackRespChanMap.Load(key); ok {
		defer m.ackRespChanMap.Delete(key)
		if ackRespChan, ok := value.(ackRespChan); ok {
			var timer *time.Timer
			var timerChan <-chan time.Time
			if timeout > 0 {
				timer = time.NewTimer(timeout)
				timerChan = timer.C
			}
			select {
			case respEntity := <-ackRespChan:
				if timer != nil {
					timer.Stop()
				}
				return respEntity.data, respEntity.err
			case <-timerChan:
				return nil, AckTimeoutError
			}
		}
	}
	return nil, nil
}

// CommitAck commit specified ack transaction.
func (m *SafeAckManager) CommitAck(key interface{}, data interface{}) {

	if key == nil {
		return
	}

	if value, ok := m.ackRespChanMap.Load(key); ok {
		if ackRespChan, ok := value.(ackRespChan); ok {
			ackRespChan <- ackRespEntity{data: data, err: nil}
		}
	}
}

// NewAckManager will create a instance of default implementation of AckManage.
// The current default implementation is SafeAckManager.
func NewAckManager() AckManager {
	return &SafeAckManager{}
}
