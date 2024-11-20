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
	NewLoggerFunc           = NewZeroLogger
	DefaultWriter io.Writer = os.Stdout
)

type Logger interface {
	With(k, v any) Logger

	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	Panic(msg string, args ...any)

	Debugf(format string, args ...any)
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
	Errorf(format string, args ...any)
	Panicf(format string, args ...any)
}

func NewLogger(format, logLevel string) (Logger, error) {
	return NewLoggerFunc(format, logLevel)
}
