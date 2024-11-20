//go:build zaplog

package log

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func init() {
	NewLoggerFunc = NewZapLogger
}

const CallerLengthMax = 30

var _ Logger = (*ZapLogger)(nil)

type ZapLogger struct {
	logger *zap.SugaredLogger
}

func NewZapLogger(format string, logLevel string) (Logger, error) {
	logger, err := newZapLogger(format, logLevel)
	if err != nil {
		return nil, errors.Wrapf(err, "new zap logger failed, format: %s, log level: %s", format, logLevel)
	}
	return &ZapLogger{logger: logger.Sugar()}, nil
}

func (l *ZapLogger) With(k, v any) Logger {
	return &ZapLogger{logger: l.logger.With(k, v)}
}

func (l *ZapLogger) Debug(msg string, args ...any) {
	l.logger.Debug(msg, args)
}

func (l *ZapLogger) Info(msg string, args ...any) {
	l.logger.Info(msg, args)
}

func (l *ZapLogger) Warn(msg string, args ...any) {
	l.logger.Warn(msg, args)
}

func (l *ZapLogger) Error(msg string, args ...any) {
	l.logger.Error(msg, args)
}

func (l *ZapLogger) Panic(msg string, args ...any) {
	l.logger.Panic(msg, args)
}

func (l *ZapLogger) Debugf(format string, args ...any) {
	l.logger.Debugf(format, args...)
}

func (l *ZapLogger) Infof(format string, args ...any) {
	l.logger.Infof(format, args...)
}

func (l *ZapLogger) Warnf(format string, args ...any) {
	l.logger.Warnf(format, args...)
}

func (l *ZapLogger) Errorf(format string, args ...any) {
	l.logger.Errorf(format, args...)
}

func (l *ZapLogger) Panicf(format string, args ...any) {
	l.logger.Panicf(format, args...)
}

func newZapLogger(format string, logLevel string) (*zap.Logger, error) {
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = func(ts time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(ts.Format(TimeFieldFormat))
	}

	config.EncodeLevel = func(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(fmt.Sprintf("[%s]", level.CapitalString()))
	}

	config.EncodeCaller = func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
		trimmedPath := caller.TrimmedPath()
		if len(trimmedPath) > CallerLengthMax {
			trimmedPath = trimmedPath[len(trimmedPath)-CallerLengthMax:]
		}
		enc.AppendString(fmt.Sprintf(fmt.Sprintf("%%%ds", CallerLengthMax), trimmedPath))
	}

	decoder, err := newDecoder(format, config)
	if err != nil {
		return nil, err
	}

	level, err := zapcore.ParseLevel(logLevel)
	if err != nil {
		return nil, err
	}

	writer := DefaultWriter.(zapcore.WriteSyncer)
	logger := zap.New(
		zapcore.NewCore(decoder, writer, level),
		// zap options
		// add caller
		zap.AddCaller(),
		zap.AddCallerSkip(1),
	)
	return logger, nil
}

func newDecoder(format string, config zapcore.EncoderConfig) (zapcore.Encoder, error) {
	var decoder zapcore.Encoder
	switch format {
	case FormatJSON:
		decoder = zapcore.NewJSONEncoder(config)
	case FormatConsole, "":
		decoder = zapcore.NewConsoleEncoder(config)
	default:
		return nil, fmt.Errorf("invalid log format: %s", format)
	}
	return decoder, nil
}
