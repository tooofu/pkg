package task

import (
	"time"

	"github.com/sirupsen/logrus"
)

type limitMode int8

const (
	IgnoreMode limitMode = iota
	WaitMode
)

type Task struct {
	log *logrus.Logger

	loc *time.Location

	worker *Worker
}
