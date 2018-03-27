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

package task

import (
	"github.com/mervinkid/allspark/parallel"
	"sync"
	"time"
)

type fixedTimePolicy uint8

const (
	fixedRatePolicy  fixedTimePolicy = iota
	fixedDelayPolicy
)

// fixedTimeScheduler is the implementation of Scheduler interface with
// fixed delay and fixed rate support for task execution.
// Polices:
//  +--------+     +--------+     +--------+
//  | rand 1 | ... | rand 2 | ... | rand 3 | ...
//  +--------+     +--------+     +--------+
//           ↑_____↑        ↑_____↑
//         fixed delay    fixed delay
//
//  +--------+     +--------+     +--------+
//  | rand 1 | ... | rand 2 | ... | rand 3 | ...
//  +--------+     +--------+     +--------+
//  ↑______________↑______________↑
//            fixed rate
// State:
//  +-----+           +---------+          +--------+
//  | NEW | → Start → | RUNNING | → Stop → | FINISH |
//  +-----+           +---------+          +--------+
type fixedTimeScheduler struct {
	// Props
	FixedTime time.Duration
	Policy    fixedTimePolicy
	Task      func()
	// State
	state       state
	stateMutex  sync.RWMutex
	scheduler   parallel.Goroutine
	commandChan commandChan
}

// Start will start scheduler for task scheduling execution.
func (s *fixedTimeScheduler) Start() error {
	s.stateMutex.Lock()
	defer s.stateMutex.Unlock()
	if s.state != stateNew {
		return nil
	}
	if s.Task == nil {
		return NoTaskError
	}

	s.commandChan = initCommandChan()

	s.scheduler = parallel.NewGoroutine(func() {
		timer := time.NewTimer(0)
		for {
			select {
			case <-s.commandChan:
				timer.Stop()
				close(s.commandChan) // Close command channel.
				return
			case <-timer.C:
				// Execute task with policy.
				executeTaskWithFixedTimePolicy(s.Policy, s.Task)
			}
			// Update timer
			timer = time.NewTimer(s.FixedTime)
		}
	})
	s.scheduler.Start()
	s.state = stateRunning

	return nil
}

// Stop will stop scheduler.
func (s *fixedTimeScheduler) Stop() {
	s.stateMutex.Lock()
	defer s.stateMutex.Unlock()
	if s.state == stateRunning {
		s.commandChan <- commandStop
		s.scheduler = nil
		s.state = stateFinish
	}
}

// IsRunning returns true is scheduler current running.
func (s *fixedTimeScheduler) IsRunning() bool {
	s.stateMutex.RLock()
	defer s.stateMutex.RUnlock()
	return s.state == stateRunning
}

// executeTaskWithFixedTimePolicy will execute specified task function with policy.
// If the policy is FixedDelay then execute in current goroutine or start a new
// goroutine for task execution.
func executeTaskWithFixedTimePolicy(policy fixedTimePolicy, task func()) {
	if task != nil {
		executor := parallel.NewGoroutine(task)
		executor.Start()
		switch policy {
		case fixedDelayPolicy:
			executor.Join()
			return
		case fixedRatePolicy:
			return
		}
	}
}
