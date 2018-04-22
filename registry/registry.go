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

package registry

import (
	"errors"
	"github.com/mervinkid/matcha/misc"
	"github.com/mervinkid/matcha/util"
)

var (
	ErrInvalidAppId        = errors.New("invalid app id")
	ErrInvalidHost         = errors.New("invalid host of url")
	ErrInvalidPort         = errors.New("invalid port of url")
	ErrUnsupportedProtocol = errors.New("invalid protocol of url")
)

type ElectionEvent uint8

const (
	MasterTake ElectionEvent = iota
	MasterLose
)

type Role uint8

const (
	Slaver Role = iota
	Master
)

type Config struct {
	AppId  string
	NodeId string
	Url    util.URL
	// Election is the callback method which will be invoked while election event happened.
	Election func(event ElectionEvent, masterId string)
}

type Registry interface {
	misc.Lifecycle
	misc.Sync
	misc.Type
}

func NewRegister(config Config) (Registry, error) {
	if config.AppId == "" {
		return nil, ErrInvalidAppId
	}
	if err := validateUrl(config.Url); err != nil {
		return nil, err
	}
	switch config.Url.Protocol {
	case "redis":
		registry := &redisRegistry{config: config}
		return registry, nil
	default:
		return nil, ErrUnsupportedProtocol
	}
}

func validateUrl(url util.URL) error {
	// Check host
	if url.Host == "" {
		return ErrInvalidHost
	}
	// Check port
	if url.Port == 0 {
		return ErrInvalidPort
	}
	return nil
}
