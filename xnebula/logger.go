package xnebula

import (
	"github.com/sirupsen/logrus"
	nebula "github.com/vesoft-inc/nebula-go/v3"
)

var _ nebula.Logger = NBLogger{}

type NBLogger struct {
	log *logrus.Logger
}

func NewLogger(log *logrus.Logger) (l *NBLogger) {
	return &NBLogger{log: log}
}

func (l NBLogger) Error(msg string) {
	l.log.Errorf(msg)
}

func (l NBLogger) Fatal(msg string) {
	l.log.Fatalf(msg)
}

func (l NBLogger) Info(msg string) {
	l.log.Infof(msg)
}

func (l NBLogger) Warn(msg string) {
	l.log.Warnf(msg)
}
