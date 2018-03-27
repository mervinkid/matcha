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

package misc

import "errors"

// LifecycleStart check and start specified Lifecycle instance.
func LifecycleStart(l Lifecycle) error {
	if l != nil {
		return l.Start()
	}
	return errors.New("lifecycle is nil")
}

// LifecycleStop check and stop specified Lifecycle instance.
func LifecycleStop(l Lifecycle) {
	if l != nil {
		l.Stop()
	}
}

// LifecycleCheckRun returns true is specified Lifecycle instance is valid and current is running.
func LifecycleCheckRun(l Lifecycle) bool {
	return l != nil && l.IsRunning()
}

// Sync invoke sync method if specified Sync instance is valid.
func SynchronizeIt(s Sync) {
	if s != nil {
		s.Sync()
	}
}

// TryClose check and close if possible.
func TryClose(c Close) {
	if c != nil {
		c.Close()
	}
}
