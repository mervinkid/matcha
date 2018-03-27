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

package logging

type Level uint8

const (
	LTrace Level = 1
	LDebug       = LTrace<<1 + 1
	LInfo        = LDebug<<1 + 1
	LWarn        = LInfo<<1 + 1
	LError       = LWarn<<1 + 1
	LNone        = LError<<1 + 1
)

type Logger interface {
	Trace(format string, args ...interface{})
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Error(format string, args ...interface{})
}

type LoggerProxy struct {
	level   Level
	loggers map[string]Logger
}

func (p *LoggerProxy) AddLogger(name string, logger Logger) {
	if name != "" && logger != nil {
		p.loggers[name] = logger
	}
}

func (p *LoggerProxy) RemoveLogger(name string) {
	if name != "" {
		delete(p.loggers, name)
	}
}

func (p *LoggerProxy) output(level Level, format string, args ...interface{}) {

	if level&p.level != p.level {
		return
	}

	for _, logger := range p.loggers {
		switch level {
		case LTrace:
			logger.Trace(format, args...)
			break
		case LDebug:
			logger.Debug(format, args...)
			break
		case LInfo:
			logger.Info(format, args...)
			break
		case LWarn:
			logger.Warn(format, args...)
			break
		case LError:
			logger.Error(format, args...)
			break
		}
	}
}

func (p *LoggerProxy) SetLevel(level Level) {
	p.level = level
}

func (p *LoggerProxy) Trace(format string, args ...interface{}) {
	p.output(LTrace, format, args...)
}

func (p *LoggerProxy) Debug(format string, args ...interface{}) {
	p.output(LDebug, format, args...)
}

func (p *LoggerProxy) Info(format string, args ...interface{}) {
	p.output(LInfo, format, args...)
}

func (p *LoggerProxy) Warn(format string, args ...interface{}) {
	p.output(LWarn, format, args...)
}

func (p *LoggerProxy) Error(format string, args ...interface{}) {
	p.output(LError, format, args...)
}

var proxy = &LoggerProxy{level: LNone, loggers: make(map[string]Logger)}

// SetLogLevel set output limit to global logger proxy.
func SetLogLevel(level Level) {
	if proxy != nil {
		proxy.SetLevel(level)
	}
}

// AddLogger register a logger into global logger proxy.
func AddLogger(name string, logger Logger) {
	if proxy != nil {
		proxy.AddLogger(name, logger)
	}
}

// RemoveLogger will cancel the specified logger from global logger proxy.
func RemoveLogger(name string) {
	if proxy != nil {
		proxy.RemoveLogger(name)
	}
}

func Trace(fmt string, args ...interface{}) {
	proxy.Trace(fmt, args...)
}

func Debug(fmt string, args ...interface{}) {
	proxy.Debug(fmt, args...)
}

func Info(fmt string, args ...interface{}) {
	proxy.Info(fmt, args...)
}

func Warn(fmt string, args ...interface{}) {
	proxy.Warn(fmt, args...)
}

func Error(fmt string, args ...interface{}) {
	proxy.Error(fmt, args...)
}
