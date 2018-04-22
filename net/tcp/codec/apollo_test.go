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

type _tUser struct {
	Id     int64
	Name   string
	Gender string
	Group  _tGroup
}

func (u *_tUser) TypeCode() uint16 {
	return 1
}

type _tGroup struct {
	Id   int64
	Name string
}

func (u *_tGroup) TypeCode() uint16 {
	return 2
}

func TestApolloFrameCodec(t *testing.T) {

	// Prepare codec
	config := ApolloConfig{}
	config.RegisterEntity(func() ApolloEntity {
		return &_tUser{}
	})
	config.RegisterEntity(func() ApolloEntity {
		return &_tGroup{}
	})
	encoder := NewApolloFrameEncoder(config)
	decoder := NewApolloFrameDecoder(config)

	// Prepare data
	user := &_tUser{}
	user.Id = 1
	user.Name = "Mervin"
	user.Gender = "M"
	group := _tGroup{}
	group.Id = 1
	group.Name = "TIG"
	user.Group = group
	t.Log("Source data:\t\t", user)

	// Encode
	encodeResult, encodeError := encoder.Encode(user)
	if encodeError != nil {
		t.Fatal(encodeError)
	}
	t.Log("Encode result:\t", encodeResult)

	// Decode
	byteBuffer := buffer.NewElasticUnsafeByteBuf(len(encodeResult))
	byteBuffer.WriteBytes(encodeResult)
	decodeResult, decodeError := decoder.Decode(byteBuffer)
	if decodeError != nil {
		t.Fatal(decodeError)
	}
	t.Log("Decode result:\t", decodeResult)

}

func BenchmarkApolloFrameEncoder_Encode(b *testing.B) {
	// Prepare codec
	config := ApolloConfig{}
	config.RegisterEntity(func() ApolloEntity {
		return &_tUser{}
	})
	config.RegisterEntity(func() ApolloEntity {
		return &_tGroup{}
	})
	encoder := NewApolloFrameEncoder(config)

	// Prepare encode source
	user := new(_tUser)
	user.Id = 1
	user.Name = "Mervin"
	user.Gender = "M"
	group := _tGroup{}
	group.Id = 1
	group.Name = "TIG"
	user.Group = group
	encodeSource := user

	// Benchmark encode
	encoder = NewApolloFrameEncoder(config)
	b.StartTimer()
	for i := 0; i < 100000; i++ {
		if _, err := encoder.Encode(encodeSource); err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
}

func BenchmarkApolloFrameDecoder_Decode(b *testing.B) {

	// Prepare codec
	config := ApolloConfig{}
	config.RegisterEntity(func() ApolloEntity {
		return &_tUser{}
	})
	config.RegisterEntity(func() ApolloEntity {
		return &_tGroup{}
	})
	encoder := NewApolloFrameEncoder(config)
	decoder := NewApolloFrameDecoder(config)

	// Prepare encode source
	user := new(_tUser)
	user.Id = 1
	user.Name = "Mervin"
	user.Gender = "M"
	group := _tGroup{}
	group.Id = 1
	group.Name = "TIG"
	user.Group = group

	// Prepare decode source
	encodeResult, encodeError := encoder.Encode(user)
	if encodeError != nil {
		b.Fatal(encodeError)
	}
	decodeSource := encodeResult

	// Benchmark decode
	b.ReportAllocs()

	b.StartTimer()
	for i := 0; i < 100000; i++ {
		byteBuffer := buffer.NewElasticUnsafeByteBuf(len(encodeResult))
		byteBuffer.WriteBytes(decodeSource)
		if _, err := decoder.Decode(byteBuffer); err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
}
