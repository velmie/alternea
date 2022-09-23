package app

type NoopLogger struct{}

func NewNoopLogger() *NoopLogger {
	return &NoopLogger{}
}

func (n *NoopLogger) WithFields(fields LogFields) Logger {
	return n
}
func (n *NoopLogger) Debug(v ...interface{})                   {}
func (n *NoopLogger) Debugf(format string, v ...interface{})   {}
func (n *NoopLogger) Error(v ...interface{})                   {}
func (n *NoopLogger) Errorf(format string, v ...interface{})   {}
func (n *NoopLogger) Info(v ...interface{})                    {}
func (n *NoopLogger) Infof(format string, v ...interface{})    {}
func (n *NoopLogger) Warning(v ...interface{})                 {}
func (n *NoopLogger) Warningf(format string, v ...interface{}) {}
func (n *NoopLogger) Level() LogLevel {
	return ErrorLevel
}
