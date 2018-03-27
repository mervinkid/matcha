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

package misc

import (
	"encoding/json"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"regexp"
	"strings"
)

var (
	regexpPropertyComment = regexp.MustCompile("#(\\w|\\W)*?(\n)")
	regexpPropertyLine    = regexp.MustCompile("^(\\w|.|-|_)+=(\\w|\\W)*")
)

// LoadPropertyFile try load configuration properties from specified property file.
func LoadPropertyFile(path string) (map[string]string, error) {
	fileContent, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	cleanContent := regexpPropertyComment.ReplaceAllString(string(fileContent), "\n")
	lines := strings.Split(cleanContent, "\n")

	config := make(map[string]string)
	for _, line := range lines {
		trimLine := strings.Trim(line, " ")
		if !regexpPropertyLine.MatchString(trimLine) {
			continue
		}
		firstEqualsIndex := strings.Index(trimLine, "=")
		var propertyKey string = line[:firstEqualsIndex]
		var propertyValue string
		if firstEqualsIndex+1 != len(trimLine) {
			propertyValue = trimLine[firstEqualsIndex+1:]
		}
		config[propertyKey] = propertyValue
	}
	return config, nil
}

// LoadJsonFile try load configuration from specified json file.
func LoadJsonFile(path string) (map[string]interface{}, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var config map[string]interface{}
	if err := json.Unmarshal(bytes, &config); err != nil {
		return nil, err
	}
	return config, nil
}

// LoadYmlFile try load configuration from specified yml file.
func LoadYmlFile(path string) (map[string]interface{}, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var config map[string]interface{}
	if err := yaml.Unmarshal(bytes, &config); err != nil {
		return nil, err
	}
	return config, nil
}
