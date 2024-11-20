package log

var l Logger

func init() {
	if err := Init(FormatConsole, LevelInfo); err != nil {
		panic(err)
	}
}

func Init(format, level string) (err error) {
	l, err = NewLogger(format, level)
	return err
}

func GetLogger() Logger {
	return l
}

func Debug(msg string, args ...any) {
	l.Debug(msg, args...)
}

func Debugf(template string, args ...any) {
	l.Debugf(template, args...)
}

func Info(msg string, args ...any) {
	l.Info(msg, args...)
}

func Infof(template string, args ...any) {
	l.Infof(template, args...)
}

func Warn(msg string, args ...any) {
	l.Warn(msg, args...)
}

func Warnf(template string, args ...any) {
	l.Warnf(template, args...)
}

func Error(msg string, args ...any) {
	l.Error(msg, args...)
}

func Errorf(template string, args ...any) {
	l.Errorf(template, args...)
}
