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

package codec

import (
	"fmt"
	"github.com/mervinkid/allspark/buffer"
)

// FrameDecoder is the interface that wraps the basic method for decode tcp stream.
// A FrameDecoder will be instantiated and init by PipelineInitializer in Pipeline
// initializing.
//
// Model:
//  +-------------------------+
//  |     (src)↓              |
//  |  Decode(in) → (result)  |
//  +-------------------------+
type FrameDecoder interface {
	Decode(in buffer.ByteBuf) (result interface{}, err error)
}

// FrameDecoder is the interface that wraps the basic method for encode tcp stream.
// A FrameEncoder will be instantiated and init by PipelineInitializer in Pipeline
// initializing.
//
// Model:
//  +-------------------------+
//  |     (src)↓              |
//  |  Encode(in) → (result)  |
//  +-------------------------+
type FrameEncoder interface {
	Encode(msg interface{}) (result []byte, err error)
}

// FrameCodec is the interface that wraps the basic method for both encode and decode.
type FrameCodec interface {
	FrameDecoder
	FrameEncoder
}

// DecodeError is a error implementation with detail error string output include
// decoder name and cause for decode exception.
// The format of complete error string is '$DECODER decode error cause  $CAUSE'.
type DecodeError struct {
	decoder string
	msg     string
}

func (e *DecodeError) Error() string {
	var prefix string
	if e.decoder != "" {
		prefix = fmt.Sprint(e.decoder, " ")
	}
	var suffix string
	if e.msg != "" {
		suffix = fmt.Sprint(" cause ", e.msg)
	}
	return fmt.Sprint(prefix, "decode error", suffix)
}

// NewDecodeError create instance of DecodeError with specified error message.
func NewDecodeError(decoder, msg string) error {
	return &DecodeError{decoder: decoder, msg: msg}
}

// EncodeError is a error implementation with detail error string output include
// encoder name and cause for encode exception.
// The format of complete error string is '$ENCODER encode error cause  $CAUSE'.
type EncodeError struct {
	encoder string
	msg     string
}

func (e *EncodeError) Error() string {
	var prefix string
	if e.encoder != "" {
		prefix = fmt.Sprint(e.encoder, " ")
	}
	var suffix string
	if e.msg != "" {
		suffix = fmt.Sprint(" cause ", e.msg)
	}
	return fmt.Sprint(prefix, "encode error", suffix)
}

// NewEncodeError create instance of EncodeError with specified error message.
func NewEncodeError(encoder, msg string) error {
	return &EncodeError{encoder: encoder, msg: msg}
}
