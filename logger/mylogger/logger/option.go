package logger

import (
	"github.com/wonksing/go-tutorials/logger/mylogger/logger/port"
	"github.com/wonksing/go-tutorials/logger/mylogger/logger/types"
)

type LoggerOption interface {
	apply(*Logger)
}

type loggerOptionFunc func(*Logger)

func (f loggerOptionFunc) apply(log *Logger) {
	f(log)
}

func WithLevel(level types.Level) LoggerOption {
	return loggerOptionFunc(func(log *Logger) {
		log.level = level
	})
}

func WithRoller(roller port.Roller) LoggerOption {
	return loggerOptionFunc(func(log *Logger) {
		log.roller = roller
	})
}

func WithStdOut(stdOut bool) LoggerOption {
	return loggerOptionFunc(func(log *Logger) {
		log.stdOut = stdOut
	})
}
