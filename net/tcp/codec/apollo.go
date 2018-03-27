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
	"encoding/binary"

	"github.com/mervinkid/allspark/buffer"
	"github.com/vmihailenco/msgpack"
)

type ApolloEntity interface {
	TypeCode() uint16
}

type ApolloConfig struct {
	TLVConfig
	entityConstructors map[uint16]func() ApolloEntity
}

func (c *ApolloConfig) RegisterEntity(constructor func() ApolloEntity) {
	c.initConfig()
	if constructor != nil {
		if testEntity := constructor(); testEntity != nil {
			c.entityConstructors[testEntity.TypeCode()] = constructor
		}
	}
}

func (c *ApolloConfig) createEntity(typeCode uint16) ApolloEntity {
	c.initConfig()
	if constructor := c.entityConstructors[typeCode]; constructor != nil {
		return constructor()
	}
	return nil
}

func (c *ApolloConfig) initConfig() {
	if c.entityConstructors == nil {
		c.entityConstructors = make(map[uint16]func() ApolloEntity)
	}
}

// ApolloFrameDecoder is a bytes to ApolloEntity decode implementation of FrameDecode based on TLVFrameDecoder
// using MessagePack for payload data deserialization.
//  +----------+-----------+---------------------------+
//  |    TAG   |  LENGTH   |           VALUE           |
//  | (1 byte) | (4 bytes) |   2 bytes   | serialized  |
//  |          |           |  type code  |    data     |
//  +----------+-----------+---------------------------+
// Decode:
//  []byte → ApolloEntity(*pointer)
type ApolloFrameDecoder struct {
	Config     ApolloConfig
	tlvDecoder FrameDecoder
}

func (d *ApolloFrameDecoder) Decode(in buffer.ByteBuf) (interface{}, error) {

	if in.ReadableBytes() == 0 {
		return d.decodeNothing()
	}

	// Decode inbound with TLVFrameDecoder
	d.initTLVDecoder()
	tlvPayload, tlvErr := d.tlvDecoder.Decode(in)
	if tlvPayload == nil && tlvErr == nil {
		return d.decodeNothing()
	}
	if tlvErr != nil {
		return d.decodeFailure(tlvErr.Error())
	}

	// Init ByteBuf for MessagePack deserialization.
	tlvPayloadByteBuffer := buffer.NewElasticUnsafeByteBuf(len(tlvPayload.([]byte)))
	tlvPayloadByteBuffer.WriteBytes(tlvPayload.([]byte))

	// Parse 2 bytes of message type code.
	if tlvPayloadByteBuffer.ReadableBytes() < 2 {
		return d.decodeFailure("illegal payload")
	}
	var typeCode uint16
	binary.Read(tlvPayloadByteBuffer, binary.BigEndian, &typeCode)

	// Parse reset bytes for serialized data.
	serializedBytes := tlvPayloadByteBuffer.ReadBytes(tlvPayloadByteBuffer.ReadableBytes())
	if entity := d.Config.createEntity(typeCode); entity != nil {
		if unmarshalErr := msgpack.Unmarshal(serializedBytes, entity); unmarshalErr != nil {
			return d.decodeFailure(unmarshalErr.Error())
		} else {
			return d.decodeSuccess(entity)
		}
	}
	return d.decodeNothing()
}

func (d *ApolloFrameDecoder) initTLVDecoder() {
	if d.tlvDecoder == nil {
		d.tlvDecoder = NewTLVFrameDecoder(d.Config.TLVConfig)
	}
}

func (d *ApolloFrameDecoder) decodeNothing() (interface{}, error) {
	return d.decodeSuccess(nil)
}

func (d *ApolloFrameDecoder) decodeSuccess(result interface{}) (interface{}, error) {
	return result, nil
}

func (d *ApolloFrameDecoder) decodeFailure(cause string) (interface{}, error) {
	return nil, NewDecodeError("ApolloFrameDecoder", cause)
}

// NewApolloFrameDecoder create a new ApolloFrameDecoder instance with configuration.
func NewApolloFrameDecoder(config ApolloConfig) FrameDecoder {
	return &ApolloFrameDecoder{Config: config}
}

// ApolloFrameEncoder is a ApolloEntity to bytes encoder implementation of FrameEncode based on TLVFrameEncoder
// using MessagePack for payload data serialization.
//  +----------+-----------+---------------------------+
//  |    TAG   |  LENGTH   |           VALUE           |
//  | (1 byte) | (4 bytes) |   2 bytes   | serialized  |
//  |          |           |  type code  |    data     |
//  +----------+-----------+---------------------------+
// Encode:
//  ApolloEntity(*pointer) → []byte
type ApolloFrameEncoder struct {
	Config     ApolloConfig
	tlvEncoder FrameEncoder
}

func (e *ApolloFrameEncoder) Encode(msg interface{}) ([]byte, error) {

	// Message must be an implementation of ApolloEntity interface.
	var entity ApolloEntity
	switch message := msg.(type) {
	case ApolloEntity:
		entity = message
	default:
		return e.encodeFailure("message is not valid implementation of ApolloEntity interface")
	}

	// Marshal entity to bytes.
	typeCode := entity.TypeCode()
	marshaledBytes, marshalErr := msgpack.Marshal(entity)
	if marshalErr != nil {
		return e.encodeFailure(marshalErr.Error())
	}
	// Build frame payload with marshaled bytes and type code.
	payloadByteBuffer := buffer.NewElasticUnsafeByteBuf(2 + len(marshaledBytes))
	binary.Write(payloadByteBuffer, binary.BigEndian, typeCode)
	binary.Write(payloadByteBuffer, binary.BigEndian, marshaledBytes)

	// Encode with TLVEncoder
	e.initTLVEncoder()
	frameBytes, encodeErr := e.tlvEncoder.Encode(payloadByteBuffer.ReadBytes(payloadByteBuffer.ReadableBytes()))
	if encodeErr != nil {
		return e.encodeFailure(encodeErr.Error())
	}

	return e.encodeSuccess(frameBytes)
}

func (e *ApolloFrameEncoder) initTLVEncoder() {
	if e.tlvEncoder == nil {
		e.tlvEncoder = NewTLVFrameEncoder(e.Config.TLVConfig)
	}
}

func (e *ApolloFrameEncoder) encodeSuccess(result []byte) ([]byte, error) {
	return result, nil
}

func (e *ApolloFrameEncoder) encodeFailure(cause string) ([]byte, error) {
	return nil, NewEncodeError("ApolloFrameEncoder", cause)
}

// NewApolloFrameEncoder create a new ApolloFrameEncoder instance with configuration.
func NewApolloFrameEncoder(config ApolloConfig) FrameEncoder {
	return &ApolloFrameEncoder{Config: config}
}
