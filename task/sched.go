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
	"errors"
	"github.com/mervinkid/allspark/misc"
	"time"
)

type state uint8

const (
	stateNew     state = iota
	stateRunning
	stateFinish
)

type command uint8

type commandChan chan command

const (
	commandStop command = iota
)

var (
	NoTaskError = errors.New("no task to be scheduled execute")
)

// Scheduler is the interface defined a scheduler for task scheduling execution.
// Methods:
//  Start will start scheduler for task scheduling execution.
//  Stop will stop scheduler.
//  IsRunning returns true is scheduler current running.
type Scheduler interface {
	misc.Lifecycle
}

// NewFixedDelayScheduler create a new scheduler instance which execute task with fixed delay time.
// Work mode:
//  +--------+     +--------+     +--------+
//  | rand 1 | ... | rand 2 | ... | rand 3 | ...
//  +--------+     +--------+     +--------+
//           ↑_____↑        ↑_____↑
//         fixed delay    fixed delay
// State:
//  +-----+           +---------+          +--------+
//  | NEW | → Start → | RUNNING | → Stop → | FINISH |
//  +-----+           +---------+          +--------+
func NewFixedDelayScheduler(task func(), delay time.Duration) Scheduler {
	return &fixedTimeScheduler{
		Task:      task,
		FixedTime: delay,
		Policy:    fixedDelayPolicy,
	}
}

// NewFixedRateScheduler create a new scheduler instance which execute task with fixed rate.
// Work mode:
//  +--------+     +--------+     +--------+
//  | rand 1 | ... | rand 2 | ... | rand 3 | ...
//  +--------+     +--------+     +--------+
//  ↑______________↑______________↑
//            fixed rate
// State:
//  +-----+           +---------+          +--------+
//  | NEW | → Start → | RUNNING | → Stop → | FINISH |
//  +-----+           +---------+          +--------+
func NewFixedRateScheduler(task func(), rate time.Duration) Scheduler {
	return &fixedTimeScheduler{
		Task:      task,
		FixedTime: rate,
		Policy:    fixedRatePolicy,
	}
}

// NewCornScheduler create a new scheduler instance with corn expression support.
func NewCornScheduler(corn string, task func()) Scheduler {
	return &cornScheduler{
		Task:    task,
		CornExp: corn,
	}
}

func initCommandChan() commandChan {
	return make(chan command, 1)
}
