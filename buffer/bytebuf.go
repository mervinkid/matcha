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

package buffer

import (
	"io"
)

// ByteBuf is the interface provide necessary method of byte buffer with double indexes.
// The implementation must be designed to consist of two indexes that are read and written separately.
//
// The relationship between reading index and writing index:
//  +--------------+--------------------+------------------+
//  |  read bytes  |   readable bytes   |  writable bytes  |
//  |              |  (unread content)  |      (free)      |
//  +--------------+--------------------+------------------+
//  ↑              ↑                    ↑                  ↑
//  0    <=   readerIndex    <=    writerIndex    <=    capacity
//  +------------------------------------------------------+
//
// Method Reset will reset both read index and write index to 0.
// Method Release will release read bytes and recalculate indexes.
type ByteBuf interface {
	io.Writer
	io.Reader

	ReadIndex() int
	ReadBytes(length int) []byte
	ReadableBytes() int

	WriteIndex() int
	WriteBytes(bytes []byte)
	WritableBytes() int

	Capacity() int
	Reset()
	Release()
}

// ElasticUnsafeByteBuf is a unsafe implementation of ByteBuf interface with elastic memory usage.
// Note:
// This implementation is not parallel safe.
type elasticUnsafeByteBuf struct {
	buffer     []byte
	readIndex  int
	writeIndex int
	capacity   int
}

// ReadBytes transfers this buffer's data to a newly created buffer starting at
// the current read index and increases the read index
// by the number of the transferred bytes.
func (pb *elasticUnsafeByteBuf) ReadBytes(length int) []byte {

	if length < 0 {
		return []byte{}
	}

	result := make([]byte, length)

	var targetReadIndex int
	if (pb.readIndex + length) <= pb.writeIndex {
		targetReadIndex = pb.readIndex + length
	} else {
		targetReadIndex = pb.writeIndex
	}

	// Copy data
	copy(result, pb.buffer[pb.readIndex:(targetReadIndex)])

	// Update indexes
	pb.readIndex = targetReadIndex

	return result
}

// WriteBytes transfers the specified source array's data to this buffer starting at the current
// write index and increases the write index by the number of the transferred bytes.
func (pb *elasticUnsafeByteBuf) WriteBytes(bytes []byte) {

	writeSize := len(bytes)

	if writeSize == 0 {
		return
	}

	// Check capacity
	if pb.calcWritableBytes() < writeSize {
		// Allocate
		newSize := int(float32(writeSize+pb.writeIndex) * 1.2)
		newBuffer := make([]byte, newSize)
		// Merge buffer
		copy(newBuffer, pb.buffer)
		pb.buffer = newBuffer
		pb.capacity = newSize
	}

	// Merge data
	copy(pb.buffer[pb.writeIndex:], bytes)
	pb.writeIndex += writeSize

}

// ReadableBytes returns the number of readable bytes
func (pb *elasticUnsafeByteBuf) ReadableBytes() int {
	return pb.calcReadableBytes()
}

func (pb *elasticUnsafeByteBuf) calcReadableBytes() int {
	return pb.writeIndex - pb.readIndex
}

// WritableBytes returns the number of writable bytes。
func (pb *elasticUnsafeByteBuf) WritableBytes() int {
	return pb.calcWritableBytes()
}

func (pb *elasticUnsafeByteBuf) calcWritableBytes() int {
	return pb.capacity - pb.writeIndex
}

// ReadIndex returns value of read index.
func (pb *elasticUnsafeByteBuf) ReadIndex() int {
	return pb.readIndex
}

// WriteIndex returns value of write index.
func (pb *elasticUnsafeByteBuf) WriteIndex() int {
	return pb.writeIndex
}

// Capacity returns capacity size with integer value.
func (pb *elasticUnsafeByteBuf) Capacity() int {
	return pb.capacity
}

// Reset will reset both read index and write index to 0.
func (pb *elasticUnsafeByteBuf) Reset() {
	pb.writeIndex = 0
	pb.readIndex = 0
}

func (pb *elasticUnsafeByteBuf) Write(p []byte) (n int, err error) {

	// Write
	pb.WriteBytes(p)
	return len(p), nil
}

func (pb *elasticUnsafeByteBuf) Read(p []byte) (n int, err error) {

	var readSize int
	if len(p) < pb.calcReadableBytes() {
		readSize = len(p)
	} else {
		readSize = pb.calcReadableBytes()
	}
	if readSize == 0 {
		return 0, io.EOF
	}
	result := pb.ReadBytes(readSize)
	copy(p, result)

	return readSize, nil
}

// Release will release read bytes and recalculate indexes.
func (pb *elasticUnsafeByteBuf) Release() {

	newBuffer := make([]byte, pb.capacity-pb.readIndex)
	copy(newBuffer, pb.buffer[pb.readIndex:pb.writeIndex])
	pb.buffer = newBuffer
	pb.writeIndex = pb.writeIndex - pb.readIndex
	pb.readIndex = 0
}

// Create a new instance of ElasticUnsafeByteBuf with init size.
func NewElasticUnsafeByteBuf(initSize int) ByteBuf {
	if initSize < 0 {
		initSize = 0
	}
	return &elasticUnsafeByteBuf{
		buffer:     make([]byte, initSize),
		readIndex:  0,
		writeIndex: 0,
		capacity:   initSize,
	}
}
