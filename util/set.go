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
	// Clear removes all of the elements from this set.
	Clear()
	// Intersection returns a Set with intersection elements between this Set and specified Set.
	Intersection(set Set) Set
	// Union returns a Set with union elements between this Set and specified Set.
	Union(set Set) Set
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

// Union returns a Set with union elements between this Set and specified Set.
func (s *safeHashSet) Union(set Set) Set {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	newSet := &safeHashSet{}
	s.setMap.Range(func(key, value interface{}) bool {
		newSet.Add(key)
		return true
	})
	if set != nil {
		set.Range(func(element interface{}) bool {
			if !newSet.Contains(element) {
				newSet.Add(element)
			}
			return true
		})
	}
	return newSet
}

// Intersection returns a Set with intersection elements between this Set and specified Set.
func (s *safeHashSet) Intersection(set Set) Set {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	newSet := &safeHashSet{}
	if set != nil {
		set.Range(func(element interface{}) bool {
			if _, ok := s.setMap.Load(element); ok {
				newSet.Add(element)
			}
			return true
		})
	}
	return newSet
}

// Clear removes all of the elements from this set.
func (s *safeHashSet) Clear() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.setMap = sync.Map{}
}

// NewSafeHashSet create a instance of Set with parallel safe support.
func newSafeHashSet() Set {
	return &safeHashSet{}
}

type hashSetMap map[interface{}]bool

// HashSet is an implementation of Set interface based on hash table.
type hashSet struct {
	setMap hashSetMap
}

// Add the specified element to this set if it is not already present.
func (s *hashSet) Add(element interface{}) {
	s.checkInit()
	_, ok := s.setMap[element]
	if !ok {
		s.setMap[element] = true
	}
}

// Remove the specified element from this set if it is present.
func (s *hashSet) Remove(element interface{}) {
	s.checkInit()
	delete(s.setMap, element)
}

// Contains returns true if this set contains the specified element.
func (s *hashSet) Contains(element interface{}) bool {
	s.checkInit()
	_, ok := s.setMap[element]
	return ok
}

// IsEmpty returns true if this set contains no elements.
func (s *hashSet) IsEmpty() bool {
	return len(s.setMap) == 0
}

// Size returns the number of elements in this set.
func (s *hashSet) Size() int {
	return len(s.setMap)
}

// Range calls f sequentially for each key and value present in the set.
// If f returns false, range stops the iteration.
func (s *hashSet) Range(f func(element interface{}) bool) {
	if f != nil {
		for k := range s.setMap {
			if !f(k) {
				break
			}
		}
	}
}

// Clear removes all of the elements from this set.
func (s *hashSet) Clear() {
	s.setMap = make(hashSetMap)
}

// Intersection returns a Set with intersection elements between this Set and specified Set.
func (s *hashSet) Intersection(set Set) Set {
	newSet := newHashSet()
	if set != nil {
		set.Range(func(element interface{}) bool {
			if _, ok := s.setMap[element]; ok {
				newSet.Add(element)
			}
			return true
		})
	}
	return newSet
}

// Union returns a Set with union elements between this Set and specified Set.
func (s *hashSet) Union(set Set) Set {
	newSet := newHashSet()
	for k := range s.setMap {
		newSet.Add(k)
	}
	if set != nil {
		set.Range(func(element interface{}) bool {
			if !newSet.Contains(element) {
				newSet.Add(element)
			}
			return true
		})
	}
	return newSet
}

func (s *hashSet) checkInit() {
	if s.setMap == nil {
		s.setMap = make(hashSetMap)
	}
}

// NewHashSet create a instance of HashSet.
func newHashSet() Set {
	return &hashSet{}
}

// NewSet create a new instance of Set.
// If the safe parameter is true, returns a new instance of SafeHashSet, or HashSet.
func NewSet(safe bool) Set {
	if safe {
		return newSafeHashSet()
	}
	return newHashSet()
}
