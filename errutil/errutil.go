package errutil

import (
	"fmt"
	"runtime"
	"strings"
)

const (
	maxFrames   = 16
	skipFrames  = 4
	fmtLogError = `Error:          %v`
	fmtLogStack = `Stack:          %v`
)

func frameStack() []string {
	frames := collectFrames()
	lines := make([]string, 0, len(frames))

	for _, frame := range frames {
		lines = append(lines, fmt.Sprintf("%s@%s:%d", frame.Function, frame.File, frame.Line))
	}
	return lines
}

func collectFrames() []runtime.Frame {
	pc := make([]uintptr, maxFrames)
	n := runtime.Callers(skipFrames, pc)
	if n == 0 {
		return nil
	}

	pc = pc[:n]
	frames := runtime.CallersFrames(pc)

	collectedFrames := make([]runtime.Frame, 0, maxFrames)
	discardedFrames := make([]runtime.Frame, 0, maxFrames)
	for {
		frame, more := frames.Next()

		// collect all frames except those from upper/db and runtime stack
		if (strings.Contains(frame.Function, "zeus/errutil") || strings.Contains(frame.Function, "/go/src/")) && !strings.Contains(frame.Function, "test") {
			discardedFrames = append(discardedFrames, frame)
		} else {
			collectedFrames = append(collectedFrames, frame)
		}

		if !more {
			break
		}
	}

	if len(collectedFrames) < 1 {
		return discardedFrames
	}

	return collectedFrames
}

// ErrorStack error stack
func ErrorStack(err error) string {
	lines := make([]string, 0, 8)

	if err != nil {
		lines = append(lines, fmt.Sprintf(fmtLogError, err))
	}

	if stack := frameStack(); len(stack) > 0 {
		lines = append(lines, fmt.Sprintf(fmtLogStack, "\n\t"+strings.Join(stack, "\n\t")))
	}

	return "\t" + strings.Replace(strings.Join(lines, "\n"), "\n", "\n\t", -1) + "\n\n"
}
