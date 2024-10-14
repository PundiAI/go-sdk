package log

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

func init() {
	zerolog.TimeFieldFormat = TimeFieldFormat
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	zerolog.CallerSkipFrameCount = 3
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		if parts := strings.Split(file, "/"); len(parts) > 2 && parts[0] == "github.com" {
			return strings.Join(parts[3:], "/") + ":" + strconv.Itoa(line)
		}
		return file + ":" + strconv.Itoa(line)
	}
}

var _ Logger = (*ZeroLogger)(nil)

type ZeroLogger struct {
	logger *zerolog.Logger
}

func NewZeroLogger(format, logLevel string) (Logger, error) {
	logger, err := newZeroLogger(format, logLevel)
	if err != nil {
		return nil, errors.Wrapf(err, "new zero logger failed, format: %s, log level: %s", format, logLevel)
	}
	return &ZeroLogger{logger: logger}, nil
}

func (l *ZeroLogger) With(k, v interface{}) Logger {
	logger := l.logger.With().Fields([]interface{}{k, v}).Logger()
	return &ZeroLogger{logger: &logger}
}

func (l *ZeroLogger) Debug(msg string, args ...interface{}) {
	l.logger.Debug().Fields(args).Msg(msg)
}

func (l *ZeroLogger) Info(msg string, args ...interface{}) {
	l.logger.Info().Fields(args).Msg(msg)
}

func (l *ZeroLogger) Warn(msg string, args ...interface{}) {
	l.logger.Warn().Fields(args).Msg(msg)
}

func (l *ZeroLogger) Error(msg string, args ...interface{}) {
	l.logger.Error().Fields(args).Msg(msg)
}

func (l *ZeroLogger) Panic(msg string, args ...interface{}) {
	l.logger.Panic().Fields(args).Msg(msg)
}

func (l *ZeroLogger) Debugf(format string, args ...interface{}) {
	l.logger.Debug().Msgf(format, args...)
}

func (l *ZeroLogger) Infof(format string, args ...interface{}) {
	l.logger.Info().Msgf(format, args...)
}

func (l *ZeroLogger) Warnf(format string, args ...interface{}) {
	l.logger.Warn().Msgf(format, args...)
}

func (l *ZeroLogger) Errorf(format string, args ...interface{}) {
	l.logger.Error().Msgf(format, args...)
}

func (l *ZeroLogger) Panicf(format string, args ...interface{}) {
	l.logger.Panic().Msgf(format, args...)
}

func newZeroLogger(format, logLevel string) (*zerolog.Logger, error) {
	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		return nil, err
	}
	zerolog.SetGlobalLevel(level)
	var w io.Writer
	switch format {
	case FormatJSON:
		w = DefaultWriter
	case FormatConsole, "":
		w = zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
			w.Out = DefaultWriter
			w.TimeFormat = TimeFieldFormat
		})
	default:
		return nil, fmt.Errorf("invalid log format: %s", format)
	}
	logger := zerolog.New(w).With().Timestamp().Caller().Logger()
	return &logger, nil
}
