package log

var _ Logger = (*nopLogger)(nil)

func NewNopLogger() Logger {
	return &nopLogger{}
}

type nopLogger struct{}

func (n *nopLogger) With(_, _ any) Logger {
	return n
}

func (*nopLogger) Debug(_ string, _ ...any) {}

func (*nopLogger) Info(_ string, _ ...any) {}

func (*nopLogger) Warn(_ string, _ ...any) {}

func (*nopLogger) Error(_ string, _ ...any) {}

func (*nopLogger) Panic(_ string, _ ...any) {}

func (*nopLogger) Debugf(_ string, _ ...any) {}

func (*nopLogger) Infof(_ string, _ ...any) {}

func (*nopLogger) Warnf(_ string, _ ...any) {}

func (*nopLogger) Errorf(_ string, _ ...any) {}

func (*nopLogger) Panicf(_ string, _ ...any) {}
