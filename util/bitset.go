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

import (
	"fmt"
)

// BitSet is the interface wraps method for BitSet data structure implementation.
type BitSet interface {
	// Clear used for set the bit specified by the index to false.
	Clear(index int)
	// Set used for set the bit at the specified index to true.
	Set(index int)
	// Get returns the value of the bit with the specified index.
	// The value is true if the bit with the index is currently set in this BitSet;
	// otherwise, the result is false.
	Get(index int) bool
	// IsEmpty returns true if this BitSet contains no bits that are set to true.
	IsEmpty() bool
	// Reset clean all bit.
	Reset()
}

// ByteSliceBitSet is a implementation of BitSet interface based on byte slice.
type byteSliceBitSet struct {
	bytes     []byte
	wordInUse int
}

func (bs *byteSliceBitSet) String() string {
	return fmt.Sprintf("byteSliceBitSet{%8b}", bs.bytes)
}

// Clear used for set the bit specified by the index to false.
func (bs *byteSliceBitSet) Clear(index int) {
	if index < 0 {
		return
	}
	// Check capacity
	if !bs.checkCapacity(index) {
		return
	}
	bs.checkAndIncreaseCapacity(index)
	// Locate byte and bit
	byteIndex, bitIndex := bs.locateBit(index)
	// Validate word is use
	if bs.bytes[byteIndex]&byte(1<<byte(bitIndex)) != 0 {
		// Decrease word in use counter
		bs.wordInUse -= 1
	}
	// Set value
	bs.bytes[byteIndex] = bs.bytes[byteIndex] & ^(1 << byte(bitIndex))
}

// Set used for set the bit at the specified index to true.
func (bs *byteSliceBitSet) Set(index int) {
	if index < 0 {
		return
	}
	// Check capacity
	bs.checkAndIncreaseCapacity(index)
	// Locate byte and bit
	byteIndex, bitIndex := bs.locateBit(index)
	// Set value
	bs.bytes[byteIndex] = bs.bytes[byteIndex] | (1 << byte(bitIndex))
	// Increase word in use counter
	bs.wordInUse += 1
}

// Get returns the value of the bit with the specified index.
// The value is true if the bit with the index is currently set in this BitSet;
// otherwise, the result is false.
func (bs *byteSliceBitSet) Get(index int) bool {
	if index < 0 {
		return false
	}
	// Check capacity
	if !bs.checkCapacity(index) {
		return false
	}
	// Local byte and bit
	byteIndex, bitIndex := bs.locateBit(index)
	// Get value
	mask := byte(1 << byte(bitIndex))
	return (bs.bytes[byteIndex] & mask) != 0
}

// IsEmpty returns true if this BitSet contains no bits that are set to true.
func (bs *byteSliceBitSet) IsEmpty() bool {
	return bs.wordInUse == 0
}

// Reset clean all bit.
func (bs *byteSliceBitSet) Reset() {
	bs.wordInUse = 0
	bs.bytes = []byte{}
}

func (bs *byteSliceBitSet) checkAndIncreaseCapacity(index int) {

	if index < 0 {
		return
	}

	if !bs.checkCapacity(index) {
		var newCapacity int
		if (index+1)%8 == 0 {
			newCapacity = (index + 1) / 8
		} else {
			newCapacity = (index+1)/8 + 1
		}
		newBytes := make([]byte, newCapacity)
		copy(newBytes, bs.bytes)
		bs.bytes = newBytes
	}
}

func (bs *byteSliceBitSet) checkCapacity(index int) bool {

	return !(cap(bs.bytes)*8-1 < index)
}

func (bs *byteSliceBitSet) locateBit(index int) (byteIndex, bitIndex int) {

	if (index+1)%8 == 0 {
		byteIndex = (index+1)/8 - 1
		bitIndex = 7
	} else {
		byteIndex = (index + 1) / 8
		bitIndex = (index+1)%8 - 1
	}

	return
}

// NewByteSliceBitSet create a new instance of byteSliceBitSet.
func NewByteSliceBitSet() BitSet {
	return &byteSliceBitSet{}
}
