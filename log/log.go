package log

import (
	"io"
	"os"
)

const (
	LevelTrace = "trace"
	LevelDebug = "debug"
	LevelInfo  = "info"
	LevelWarn  = "warn"
	LevelError = "error"
	LevelPanic = "panic"
	LevelFatal = "fatal"

	FormatConsole = "console"
	FormatJSON    = "json"

	TimeFieldFormat = "2006-01-02T15:04:05"
)

var (
	NewLoggerFunc newLogger = NewZeroLogger
	DefaultWriter io.Writer = os.Stdout
)

type newLogger func(format, logLevel string) (Logger, error)

type Logger interface {
	With(k, v interface{}) Logger

	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Panic(msg string, args ...interface{})

	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Panicf(format string, args ...interface{})
}

func NewLogger(format, logLevel string) (Logger, error) {
	return NewLoggerFunc(format, logLevel)
}
