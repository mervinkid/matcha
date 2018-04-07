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

package util_test

import (
	"github.com/mervinkid/allspark/util"
	"testing"
)

var (
	sampleA = []int{
		1, 2, 3, 4, 5, 6, 7, 8, 9, 0,
	}
	sampleB = []int{
		11, 12, 13, 14, 15, 16, 17, 18, 19, 10,
	}
)

func TestSafeHashSet(t *testing.T) {
	testSet(t, true)
}

func TestHashSet(t *testing.T) {
	testSet(t, false)
}

func testSet(t *testing.T, safe bool) {
	set := util.NewSet(safe)
	for _, item := range sampleA {
		set.Add(item)
	}
	for _, item := range sampleA {
		if !set.Contains(item) {
			t.Fail()
		}
	}
	for _, item := range sampleB {
		if set.Contains(item) {
			t.Fail()
		}
	}
}
