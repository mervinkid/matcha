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

package util

import "sync"

// Set is a interface defined a collection that contains no duplicate elements.
type Set interface {
	// Add the specified element to this set if it is not already present.
	Add(element interface{})
	// Remove the specified element from this set if it is present.
	Remove(element interface{})
	// Contains returns true if this set contains the specified element.
	Contains(element interface{}) bool
	// IsEmpty returns true if this set contains no elements.
	IsEmpty() bool
	// Size returns the number of elements in this set.
	Size() int
	// Range calls f sequentially for each key and value present in the set.
	// If f returns false, range stops the iteration.
	Range(f func(element interface{}) bool)
}

// SafeHashSet is an implementation of Set interface provide parallel safe support.
type safeHashSet struct {
	setMap   sync.Map
	mutex    sync.RWMutex
	elements int
}

// Range calls f sequentially for each key and value present in the set.
// If f returns false, range stops the iteration.
func (s *safeHashSet) Range(f func(element interface{}) bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	s.setMap.Range(func(key, _ interface{}) bool {
		return f(key)
	})
}

// Add the specified element to this set if it is not already present.
func (s *safeHashSet) Add(element interface{}) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	_, loaded := s.setMap.Load(element)
	if !loaded {
		s.setMap.Store(element, true)
		s.elements++
	}
}

// Remove the specified element from this set if it is present.
func (s *safeHashSet) Remove(element interface{}) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	_, ok := s.setMap.Load(element)
	if ok {
		s.setMap.Delete(element)
		s.elements--
	}
}

// Contains returns true if this set contains the specified element.
func (s *safeHashSet) Contains(element interface{}) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	_, ok := s.setMap.Load(element)
	return ok
}

// IsEmpty returns true if this set contains no elements.
func (s *safeHashSet) IsEmpty() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.elements == 0
}

// Size returns the number of elements in this set.
func (s *safeHashSet) Size() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.elements
}

// NewSafeHashSet create a instance of Set with parallel safe support.
func NewSafeHashSet() Set {
	return &safeHashSet{}
}
