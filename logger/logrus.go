package logger

import (
	"fmt"
	"os"
	"path"
	"runtime"

	"github.com/sirupsen/logrus"
)

// ConsoleLogger console logrus logger
func ConsoleLogger(caller bool, level logrus.Level) (*logrus.Logger, error) {
	logger := logrus.New()

	// 1. output
	logger.SetOutput(os.Stdout)

	// 2. formatter
	if caller {
		logger.SetReportCaller(true)
		// logger.SetFormatter(&logrus.JSONFormatter{
		logger.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: "2006-01-02 15:03:04",

			CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
				funcName := path.Base(frame.Function)
				fileName := fmt.Sprintf("%s:%d", path.Base(frame.File), frame.Line)
				return funcName, fileName
			},
		})
	} else {
		// logger.SetFormatter(&logrus.JSONFormatter{
		logger.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: "2006-01-02 15:03:04",
			// ForceColors: true,
			DisableQuote: true,
		})
	}

	// 3. level
	logger.SetLevel(level)

	return logger, nil
}
