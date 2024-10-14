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

func Debug(msg string, args ...interface{}) {
	l.Debug(msg, args...)
}

func Debugf(template string, args ...interface{}) {
	l.Debugf(template, args...)
}

func Info(msg string, args ...interface{}) {
	l.Info(msg, args...)
}

func Infof(template string, args ...interface{}) {
	l.Infof(template, args...)
}

func Warn(msg string, args ...interface{}) {
	l.Warn(msg, args...)
}

func Warnf(template string, args ...interface{}) {
	l.Warnf(template, args...)
}

func Error(msg string, args ...interface{}) {
	l.Error(msg, args...)
}

func Errorf(template string, args ...interface{}) {
	l.Errorf(template, args...)
}
