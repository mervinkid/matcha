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

package bind

import (
	"errors"
	"net"
	"sync"

	"github.com/mervinkid/matcha/logging"
	"github.com/mervinkid/matcha/misc"
	"github.com/mervinkid/matcha/parallel"
)

var NilListenerError = errors.New("listener is nil")
var NilCallbackError = errors.New("callback is nil")

// Acceptor is a interface wraps necessary methods for network connection acceptance.
// The implementation should be based on FSM.
type Acceptor interface {
	misc.Lifecycle
	misc.Sync
}

// AcceptorProp is a data struct for acceptor initialization.
type AcceptorProp struct {
	Parallelism    uint8
	Listener       net.Listener
	AcceptCallback func(conn net.Conn)
}

// ParallelAcceptor is a implementation of Acceptor which provide connection parallel acceptance.
// After a new connection have been accepted, the accept callback function which user defined
// in AcceptorProp will be invoked by acceptor goroutine.
type parallelAcceptor struct {
	// Props
	prop AcceptorProp
	// State
	running        bool
	stateMutex     sync.RWMutex
	stateWaitGroup sync.WaitGroup
	workerCounter  uint8
}

// Start only work on acceptor is not running. It will start goroutines for connection
// parallel acceptance.
func (pa *parallelAcceptor) Start() error {

	if pa.prop.Listener == nil {
		return NilListenerError
	}

	if pa.prop.AcceptCallback == nil {
		return NilCallbackError
	}

	// Mutex
	pa.stateMutex.Lock()
	defer pa.stateMutex.Unlock()

	if pa.running {
		// Only work on acceptor is not running.
		return nil
	}

	pa.stateWaitGroup.Add(1)

	for i := uint8(0); i < pa.prop.Parallelism; i++ {
		workerIndex := i
		workerCoroutine := parallel.NewGoroutine(func() {

			logging.Trace("AcceptWorker-%d for %s start.", workerIndex, pa.prop.Listener.Addr().String())

			defer func() {
				pa.stateMutex.Lock()
				defer pa.stateMutex.Unlock()
				pa.workerCounter -= 1
				if pa.workerCounter == 0 {
					pa.running = false
					pa.stateWaitGroup.Done()
				}
				logging.Trace("AcceptWorker-%d for %s stop.", workerIndex, pa.prop.Listener.Addr().String())
			}()

			for {
				conn, err := pa.prop.Listener.(*net.TCPListener).AcceptTCP()
				if err != nil {
					return
				}
				pa.prop.AcceptCallback(conn)
			}

		})
		pa.workerCounter += 1
		workerCoroutine.Start()
	}
	pa.running = true
	return nil
}

// IsRunning returns true if acceptor is current running.
func (pa *parallelAcceptor) IsRunning() bool {
	pa.stateMutex.RLock()
	defer pa.stateMutex.RUnlock()
	return pa.running
}

// Stop will close network listener which bind with acceptor
// and stop all parallel accept goroutine.
func (pa *parallelAcceptor) Stop() {
	pa.stateMutex.Lock()
	defer pa.stateMutex.Unlock()

	if pa.running {
		pa.prop.Listener.Close()
	}
}

// Sync block invoker goroutine until acceptor stop.
func (pa *parallelAcceptor) Sync() {
	pa.stateWaitGroup.Wait()
}

// Create a new ParallelAcceptor with acceptor properties.
func NewParallelAcceptor(prop AcceptorProp) Acceptor {
	return &parallelAcceptor{prop: prop, running: false}
}
