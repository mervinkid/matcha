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
	"github.com/mervinkid/matcha/buffer"
	"testing"
)

func TestTLVCodec(t *testing.T) {

	cfg := TLVConfig{}
	cfg.TagValue = 170
	cfg.FrameLimit = 1024 * 1024 * 4

	encoder := NewTLVFrameEncoder(cfg)

	source := []byte("Hello World.")

	t.Log("Source: ", source)

	byteBuffer := buffer.NewElasticUnsafeByteBuf(1024)
	byteBuffer.WriteBytes(source)

	encodeResultBytes, err := encoder.Encode(byteBuffer.ReadBytes(byteBuffer.ReadableBytes()))

	if err != nil {
		t.Fatal(err)
	}

	t.Log("Encode: ", encodeResultBytes)

	decoder := NewTLVFrameDecoder(cfg)
	byteBuffer.Reset()
	byteBuffer.WriteBytes(encodeResultBytes)
	byteBuffer.WriteBytes(encodeResultBytes)

	for {
		result, decodeErr := decoder.Decode(byteBuffer)
		if result == nil && decodeErr == nil {
			break
		}
		if decodeErr != nil {
			t.Fatal(decodeErr)
			continue
		}
		t.Log("Decode: ", result)
	}

}
