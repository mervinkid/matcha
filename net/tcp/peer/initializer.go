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

package peer

import (
	"github.com/mervinkid/matcha/net/tcp/codec"
)

// ChannelHandler is the interface provide necessary methods for pipeline initialization which invoked by pipeline.
// Method:
//  InitDecoder used for decoder initialization.
//  InitEncoder used for encoder initialization.
//  InitHandler used for channel handler initialization.
type PipelineInitializer interface {
	InitDecoder() codec.FrameDecoder
	InitEncoder() codec.FrameEncoder
	InitHandler() ChannelHandler
}

// FunctionalPipelineInitializer is a public implementation of PipelineInitializer interface which
// support functional definition for pipeline initialization logic.
type FunctionalPipelineInitializer struct {
	DecoderInit func() codec.FrameDecoder
	EncoderInit func() codec.FrameEncoder
	HandlerInit func() ChannelHandler
}

func (i *FunctionalPipelineInitializer) InitDecoder() codec.FrameDecoder {
	if i.DecoderInit != nil {
		return i.DecoderInit()
	}
	return nil
}

func (i *FunctionalPipelineInitializer) InitEncoder() codec.FrameEncoder {
	if i.EncoderInit != nil {
		return i.EncoderInit()
	}
	return nil
}

func (i *FunctionalPipelineInitializer) InitHandler() ChannelHandler {
	if i.HandlerInit != nil {
		return i.HandlerInit()
	}
	return nil
}
