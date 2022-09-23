package app

type LogFields map[string]interface{}

// Logger abstracts application logger
type Logger interface {
	WithFields(fields LogFields) Logger
	Debug(v ...interface{})
	Debugf(format string, v ...interface{})
	Error(v ...interface{})
	Errorf(format string, v ...interface{})
	Info(v ...interface{})
	Infof(format string, v ...interface{})
	Warning(v ...interface{})
	Warningf(format string, v ...interface{})
	Level() LogLevel
}
