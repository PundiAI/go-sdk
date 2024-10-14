package log

var _ Logger = (*nopLogger)(nil)

func NewNopLogger() Logger {
	return &nopLogger{}
}

type nopLogger struct{}

func (n *nopLogger) With(_, _ interface{}) Logger {
	return n
}

func (n *nopLogger) Debug(_ string, _ ...interface{}) {}

func (n *nopLogger) Info(_ string, _ ...interface{}) {}

func (n *nopLogger) Warn(_ string, _ ...interface{}) {}

func (n *nopLogger) Error(_ string, _ ...interface{}) {}

func (n *nopLogger) Panic(_ string, _ ...interface{}) {}

func (n *nopLogger) Debugf(_ string, _ ...interface{}) {}

func (n *nopLogger) Infof(_ string, _ ...interface{}) {}

func (n *nopLogger) Warnf(_ string, _ ...interface{}) {}

func (n *nopLogger) Errorf(_ string, _ ...interface{}) {}

func (n *nopLogger) Panicf(_ string, _ ...interface{}) {}
