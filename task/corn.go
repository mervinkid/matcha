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

package task

import (
	"errors"
	"fmt"
	"github.com/mervinkid/matcha/logging"
	"github.com/mervinkid/matcha/parallel"
	"github.com/mervinkid/matcha/util"
	"math"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	InvalidCornExpressionError = errors.New("invalid corn expression")
)

// Range constants
const (
	secondMin  = 0
	secondMax  = 59
	minuteMin  = 0
	minuteMax  = 59
	hourMin    = 0
	hourMax    = 23
	dayMin     = 1
	dayMax     = 31
	monthMin   = 1
	monthMax   = 12
	weekdayMin = 0
	weekdayMax = 6
	yearMin    = -1
	yearMax    = -1
)

// Regular expressions
var (
	regexpAll, _      = regexp.Compile("^\\*$")                          // Match '*'
	regexpRange, _    = regexp.Compile("^(\\d)+-(\\d)+$")                // Match 'NUM-NUM'
	regexpDisperse, _ = regexp.Compile("^(\\d)+(,(\\d)+)*$")             // Match 'NUM,NUM,NUM'
	regexpStep, _     = regexp.Compile("^(\\*)|((\\d)+-(\\d)+)/(\\d)+$") // Match '*/NUM' and 'NUM-NUM/NUM'
)

type cornData struct {
	Seconds  util.BitSet // Seconds vector
	Minutes  util.BitSet // Minutes vector
	Hours    util.BitSet // Hours vector
	Days     util.BitSet // Days vector
	Months   util.BitSet // Months vector
	Weekdays util.BitSet // Weekdays vector
	Years    util.BitSet // Years vector
}

func initCornData() *cornData {
	return &cornData{
		Seconds:  util.NewByteSliceBitSet(),
		Minutes:  util.NewByteSliceBitSet(),
		Hours:    util.NewByteSliceBitSet(),
		Days:     util.NewByteSliceBitSet(),
		Months:   util.NewByteSliceBitSet(),
		Weekdays: util.NewByteSliceBitSet(),
		Years:    util.NewByteSliceBitSet(),
	}
}

func (d *cornData) String() string {
	return fmt.Sprintf("cornData{Seconds:%v, Minute:%v, Hour:%v, Days:%v, Months:%v, Weekdays:%v, Year:%v}",
		d.Seconds, d.Minutes, d.Hours, d.Days, d.Months, d.Weekdays, d.Years)
}

// CornScheduler is the implementation of Scheduler interface provide corn expression support.
type cornScheduler struct {
	// Props
	CornExp  string
	Task     func()
	cornData *cornData
	// State
	state      state
	stateMutex sync.RWMutex
	stopC      stopChan
}

// Start will start scheduler for task scheduling execution.
func (s *cornScheduler) Start() error {

	if s.Task == nil {
		return NoTaskError
	}

	s.stateMutex.Lock()
	defer s.stateMutex.Unlock()

	if s.state != stateNew {
		return nil
	}

	// Init corn data sheet
	parsed, parseErr := parseCornExp(s.CornExp)
	if parseErr != nil {
		return parseErr
	}
	s.cornData = parsed

	s.stopC = initStopChan()

	scheduler := parallel.NewGoroutine(func() {
		// Whole second alignment
		offset := int64(time.Second) - time.Now().UnixNano()%int64(time.Second)
		ticker := time.NewTicker(time.Duration(offset) * time.Nanosecond)
		firstExecute := true

		var latestTaskExecuteTimestamp int64
		for {
			select {
			case <-s.stopC:
				ticker.Stop()
				return
			case <-ticker.C:
				now := time.Now()
				nowUnix := now.Unix()
				if matchCornData(*s.cornData, now) && nowUnix != latestTaskExecuteTimestamp {
					logging.Trace("CornScheduler start task at %v.", now.String())
					parallel.NewGoroutine(s.Task).Start()
					latestTaskExecuteTimestamp = nowUnix
				}
			}
			// Match corn data every second
			if !firstExecute {
				firstExecute = false
				ticker = time.NewTicker(1*time.Second - 10*time.Millisecond)
			}
		}
	})
	scheduler.Start()
	s.state = stateRunning

	return nil
}

// Stop will stop scheduler.
func (s *cornScheduler) Stop() {
	s.stateMutex.Lock()
	defer s.stateMutex.Unlock()
	if s.state == stateRunning {
		close(s.stopC)
		s.state = stateFinish
	}
}

// IsRunning returns true is scheduler current running.
func (s *cornScheduler) IsRunning() bool {
	s.stateMutex.RLock()
	defer s.stateMutex.RUnlock()
	return s.state == stateRunning
}

// Check match between specified corn data and time.
func matchCornData(data cornData, time time.Time) bool {

	return matchBitSet(data.Seconds, time.Second()) &&
		matchBitSet(data.Minutes, time.Minute()) &&
		matchBitSet(data.Hours, time.Hour()) &&
		matchBitSet(data.Days, time.Day()) &&
		matchBitSet(data.Months, int(time.Month())) &&
		matchBitSet(data.Weekdays, int(time.Weekday())) &&
		matchBitSet(data.Years, time.Year())
}

// Parse specified expression to corn data.
func parseCornExp(expression string) (*cornData, error) {

	cornExpParts, err := splitCornExpression(expression)
	if err != nil {
		return nil, err
	}

	cornData := initCornData()

	// Set seconds
	if err = setBitSet(cornData.Seconds, cornExpParts[0], secondMin, secondMax); err != nil {
		return nil, err
	}
	// Set minutes
	if err = setBitSet(cornData.Minutes, cornExpParts[1], minuteMin, minuteMax); err != nil {
		return nil, err
	}
	// Set hours
	if err = setBitSet(cornData.Hours, cornExpParts[2], hourMin, hourMax); err != nil {
		return nil, err
	}
	// Set days
	if err = setBitSet(cornData.Days, cornExpParts[3], dayMin, dayMax); err != nil {
		return nil, err
	}
	// Set months
	if err = setBitSet(cornData.Months, cornExpParts[4], monthMin, monthMax); err != nil {
		return nil, err
	}
	// Set weekdays
	if err = setBitSet(cornData.Weekdays, cornExpParts[5], weekdayMin, weekdayMax); err != nil {
		return nil, err
	}
	// Set years
	if err = setBitSet(cornData.Years, cornExpParts[6], yearMin, yearMax); err != nil {
		return nil, err
	}

	return cornData, nil
}

// Split specified expression string with space and validate.
func splitCornExpression(expression string) ([]string, error) {

	if expression == "" {
		return nil, InvalidCornExpressionError
	}
	// Split parts
	cornExpParts := strings.Split(expression, " ")
	// Validate parts
	if len(cornExpParts) != 8 || !strings.Contains(cornExpParts[7], "?") {
		return nil, InvalidCornExpressionError
	}
	// Remove spaces
	for i := range cornExpParts {
		cornExpParts[i] = strings.Replace(cornExpParts[i], " ", "", -1)
	}
	return cornExpParts, nil
}

func setBitSet(target util.BitSet, exp string, min int, max int) error {

	if target == nil {
		return nil
	}

	// Match "all" rule
	if success, err := trySetBitSetAll(target, exp); err != nil || success {
		return err
	}

	// Match "range" rule
	if success, err := trySetBitSetRange(target, exp, min, max); err != nil || success {
		return err
	}

	// Match "disperse"
	if success, err := trySetBitSetDisperse(target, exp, min, max); err != nil || success {
		return err
	}

	// Match "step" rule
	if success, err := trySetBitSetStep(target, exp, min, max); err != nil || success {
		return err
	}

	return InvalidCornExpressionError
}

func trySetBitSetAll(target util.BitSet, exp string) (bool, error) {
	if regexpAll != nil && regexpAll.MatchString(exp) {
		target.Reset()
		return true, nil
	}
	return false, nil
}

func trySetBitSetRange(target util.BitSet, exp string, min int, max int) (bool, error) {
	if regexpRange != nil && regexpRange.MatchString(exp) {
		target.Reset()
		rangeParts := strings.Split(exp, "-")
		start, err := strconv.Atoi(rangeParts[0])
		if err != nil {
			return false, err
		} else {
			start = int(math.Max(float64(start), float64(min)))
		}
		end, err := strconv.Atoi(rangeParts[1])
		if err != nil {
			return false, err
		} else {
			end = int(math.Min(float64(end), float64(max)))
		}
		for i := start; i <= end; i++ {
			target.Set(i)
		}
		return true, nil
	}
	return false, nil
}

func trySetBitSetDisperse(target util.BitSet, exp string, min int, max int) (bool, error) {
	if regexpDisperse != nil && regexpDisperse.MatchString(exp) {
		target.Reset()
		disperseParts := strings.Split(exp, ",")
		for _, part := range disperseParts {
			value, err := strconv.Atoi(part)
			if err == nil {
				if value >= min && value <= max {
					target.Set(value)
				}
			}
		}
		return true, nil
	}
	return false, nil
}

func trySetBitSetStep(target util.BitSet, exp string, min int, max int) (bool, error) {
	if regexpStep != nil && regexpStep.MatchString(exp) {
		perParts := strings.Split(exp, "/")
		leftPart := perParts[0]
		rightPart := perParts[1]

		var start, end int

		if regexpAll != nil && regexpAll.MatchString(leftPart) {
			start = min
			end = max
		}

		if regexpRange != nil && regexpRange.MatchString(leftPart) {
			rangeParts := strings.Split(leftPart, "-")
			startTmp, err := strconv.Atoi(rangeParts[0])
			if err != nil {
				return false, err
			} else {
				start = int(math.Max(float64(startTmp), float64(min)))
			}
			endTmp, err := strconv.Atoi(rangeParts[1])
			if err != nil {
				return false, err
			} else {
				end = int(math.Min(float64(endTmp), float64(max)))
			}
		}

		step, err := strconv.Atoi(rightPart)
		if err != nil {
			return false, err
		}

		for i := start; i <= end; i += step {
			target.Set(i)
		}
		return true, nil
	}
	return false, nil
}

// Check match between the specified set and offset.
func matchBitSet(source util.BitSet, value int) bool {
	return source != nil && (source.Get(value) || source.IsEmpty())
}
