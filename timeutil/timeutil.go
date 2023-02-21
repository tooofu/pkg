package timeutil

import "time"

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
