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

package task_test

import (
	"github.com/mervinkid/allspark/logging"
	"github.com/mervinkid/allspark/task"
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestFixedDelayScheduler(t *testing.T) {

	mutex := new(sync.Mutex)
	cond := new(sync.Cond)
	cond.L = mutex

	logging.SetLogLevel(logging.LInfo)

	scheduler := task.NewFixedDelayScheduler(func() {
		logging.Info("Task start. Number of goroutine is %d.", runtime.NumGoroutine())
		time.Sleep(2 * time.Second)
		logging.Info("Task finish.")
	}, 1*time.Second)
	scheduler.Start()
	time.Sleep(5 * time.Second)
	scheduler.Stop()
	time.Sleep(5 * time.Second)

}

func TestFixedRateScheduler(t *testing.T) {

	mutex := new(sync.Mutex)
	cond := new(sync.Cond)
	cond.L = mutex

	logging.SetLogLevel(logging.LInfo)

	scheduler := task.NewFixedRateScheduler(func() {
		logging.Info("Task start. Number of goroutine is %d.", runtime.NumGoroutine())
		time.Sleep(2 * time.Second)
		logging.Info("Task finish.")
	}, 1*time.Second)
	scheduler.Start()
	time.Sleep(5 * time.Second)
	scheduler.Stop()
	time.Sleep(5 * time.Second)
}

func TestNewCornScheduler(t *testing.T) {

	logging.SetLogLevel(logging.LTrace)

	taskFun := func() {
		logging.Info("Work.")
	}

	scheduler := task.NewCornScheduler("* * * * * * * ?", taskFun)
	scheduler.Start()

	time.Sleep(120 * time.Second)
	scheduler.Stop()
	time.Sleep(5 * time.Second)
}
