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
	"regexp"
	"strconv"
	"strings"
)

var (
	regexpProtocol = regexp.MustCompile("^[\\w-_.]+://")
	regexpAuth     = regexp.MustCompile("^[\\w-_]+(:[\\w-_!@#$%^&*]+)*@")
	regexpHost     = regexp.MustCompile("^[\\w-_.]+")
	regexpPort     = regexp.MustCompile("^:(\\d)+")
	regexpPath     = regexp.MustCompile("^(/[\\w-_]*)+")
	regexpParam    = regexp.MustCompile("^(\\?[\\w-_]+=[\\w\\W]+)(&[\\w-_]+=[\\w\\W]+)*")
)

// URL represents a Uniform Resource Locator, a pointer to a "resource" on the World Wide Web.
type URL struct {
	Protocol string
	User     string
	Password string
	Host     string
	Port     int
	Path     string
	Param    map[string]string
}

func (url *URL) String() string {
	result := ""
	if url.Protocol != "" {
		result += url.Protocol + "://"
	}
	if url.User != "" {
		result += url.User
		if url.Password != "" {
			result += ":" + url.Password
		}
		result += "@"
	}
	if url.Host != "" {
		result += url.Host
	}
	if url.Port != 0 {
		result += ":" + strconv.Itoa(url.Port)
	}
	if url.Path != "" {
		result += url.Path
	}
	if len(url.Param) > 0 {
		result += "?"
		paramIndex := 0
		for k, v := range url.Param {
			if paramIndex != 0 {
				result += "&"
			}
			result += k + "=" + v
			paramIndex ++
		}
	}
	return result
}

// ParseUrl parse url instance from string.
func (url *URL) Parse(src string) {
	src = strings.Trim(src, " ")

	// Parse protocol
	protocolSeq := regexpProtocol.FindString(src)
	if len(protocolSeq) >= 3 {
		url.Protocol = protocolSeq[:len(protocolSeq)-3]
	} else {
		url.Protocol = ""
	}
	src = strings.Replace(src, protocolSeq, "", 1)

	// Parse auth
	authSeq := regexpAuth.FindString(src)
	if len(authSeq) > 1 {
		authParts := strings.Split(authSeq[:len(authSeq)-1], ":")
		url.User = authParts[0]
		if len(authParts) > 1 {
			url.Password = authParts[1]
		}
	}
	src = strings.Replace(src, authSeq, "", 1)

	// Parse host
	hostSeq := regexpHost.FindString(src)
	url.Host = hostSeq
	src = strings.Replace(src, hostSeq, "", 1)

	// Parse port
	portSeq := regexpPort.FindString(src)
	url.Port, _ = strconv.Atoi(strings.Replace(portSeq, ":", "", 1))
	src = strings.Replace(src, portSeq, "", 1)

	// Parse path
	pathSeq := regexpPath.FindString(src)
	url.Path = pathSeq
	src = strings.Replace(src, pathSeq, "", 1)

	// Parse params
	paramSeq := regexpParam.FindString(src)
	if len(paramSeq) > 0 {
		if url.Param == nil {
			url.Param = make(map[string]string)
		}
		for _, param := range strings.Split(paramSeq[1:], "&") {
			paramParts := strings.Split(param, "=")
			if len(paramParts) >= 2 {
				url.Param[paramParts[0]] = paramParts[1]
			}
		}
	}
}

// ParseUrl parse url data from string and returns a new URL instance.
func ParseUrl(src string) URL {
	url := URL{}
	url.Parse(src)
	return url
}
