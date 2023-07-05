package cron

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"runtime"
	"time"
)

// Locker provides a method to lock jobs from running
// at the same time on multiple instances of cron.
// You can provide any locker implemention you wish.
type Locker interface {
	Lock(key string) (bool, error)
	Unlock(key string) error
}

const MAXJOBNUM = 10000

type timeUnit int

// //go:generate stringer -type=timeUnit
const (
	seconds timeUnit = iota + 1
	minutes
	hours
	days
	weeks
	// months
	// duration
	// crontab
)

type limitMode int8

const (
	IgnoreMode limitMode = iota
	WaitMode
)

var (
	// Time location, default set by the time.Local (*time.Location)
	loc    = time.Local
	locker Locker
)

var (
	ErrInvalidInterval           = errors.New("cron: invalid interval")
	ErrInvalidIntervalType       = errors.New("cron: invalid interval type")
	ErrInvalidFunctionParameters = errors.New("cron: length of function parameters must match job function paramters")
	ErrUnsupportedTimeFormat     = errors.New("cron: unsupported time format")
)

// type nextTime struct {
//     duration time.Duration
//     dateTime time.Time
// }

func wrapOrError(toWrap error, err error) (returnErr error) {
	if toWrap != nil && !errors.Is(err, toWrap) {
		returnErr = fmt.Errorf("%s: %w", err, toWrap)
	} else {
		returnErr = err
	}
	return
}

// regex patterns for supported time formats
var (
	timeWithSeconds    = regexp.MustCompile(`(?m)^\d{1,2}:\d\d:\d\d$`)
	timeWithoutSeconds = regexp.MustCompile(`(?m)^\d{1,2}:\d\d$`)
)

func callJobFunc(jf interface{}) {
	if jf == nil {
		return
	}

	f := reflect.ValueOf(jf)
	if !f.IsZero() {
		f.Call([]reflect.Value{})
	}
}

func callJobFuncWithParams(jf interface{}, params []interface{}) (err error) {
	if jf == nil {
		return
	}

	f := reflect.ValueOf(jf)
	if f.IsZero() {
		return nil
	}

	if len(params) != f.Type().NumIn() {
		// return nil, ErrParamsNotAdapted
		return nil
	}
	in := make([]reflect.Value, len(params))
	for i, param := range params {
		in[i] = reflect.ValueOf(param)
	}

	vals := f.Call(in)
	var ok bool
	for _, val := range vals {
		i := val.Interface()
		if err, ok = i.(error); ok {
			return
		}
	}
	return
}

// func ChangeLoc(newLocation *time.Location) {
//     loc = newLocation
//     defaultScheduler.ChangeLoc(newLocation)
// }

// func SetLocker(l Locker) {
//     locker = l
// }

func getFunctionName(fn interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
}

func getFunctionKey(funcName string) string {
	h := sha256.New()
	h.Write([]byte(funcName))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func parseTime(t string) (hour, min, sec int, err error) {
	var timeLayout string
	switch {
	case timeWithSeconds.Match([]byte(t)):
		timeLayout = "15:04:05"
	case timeWithoutSeconds.Match([]byte(t)):
		timeLayout = "15:04"
	default:
		return 0, 0, 0, ErrUnsupportedTimeFormat
	}

	parsedTime, err := time.Parse(timeLayout, t)
	if err != nil {
		return 0, 0, 0, ErrUnsupportedTimeFormat
	}
	return parsedTime.Hour(), parsedTime.Minute(), parsedTime.Second(), nil
}
