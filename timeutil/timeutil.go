package timeutil

import (
	"errors"
	"regexp"
	"time"
)

var (
	timeWithSeconds    = regexp.MustCompile(`(?m)^\d{1,2}:\d\d:\d\d$`)
	timeWithoutSeconds = regexp.MustCompile(`(?m)^\d{1,2}:\d\d$`)

	ErrUnsupportedTimeFormat = errors.New("unsupported time format")
)

// UnixMilli unix milli
func UnixMilli() (milli int64) {
	// milli = time.Now().UnixMilli(); // go 1.17+
	milli = time.Now().UnixNano() / int64(time.Millisecond)
	return
}

// Now tz
func Now() (now time.Time) {
	// TODO timezone
	// l, _ := time.LoadLocation("Asia/Shanghai")
	// now := time.Now().In(l)
	now = time.Now()
	return
}

func ParseTime(t string) (hour, min, sec int, err error) {
	var timeLayout string
	switch {
	case timeWithSeconds.Match([]byte(t)):
		timeLayout = "15:04:05"
	case timeWithoutSeconds.Match([]byte(t)):
		timeLayout = "15:04"
	default:
		return 0, 0, 0, ErrUnsupportedTimeFormat
	}

	// layout := "2006-01-02 15:04:05 -0700 MST"
	parsedTime, err := time.Parse(timeLayout, t)
	if err != nil {
		return 0, 0, 0, ErrUnsupportedTimeFormat
	}
	return parsedTime.Hour(), parsedTime.Minute(), parsedTime.Second(), nil
}
