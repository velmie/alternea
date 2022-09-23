package app

import "github.com/sirupsen/logrus"

var (
	levelToLogrusLevel = map[LogLevel]logrus.Level{
		ErrorLevel: logrus.ErrorLevel,
		WarnLevel:  logrus.WarnLevel,
		InfoLevel:  logrus.InfoLevel,
		DebugLevel: logrus.DebugLevel,
	}
	logrusLevelToLevel = map[logrus.Level]LogLevel{
		logrus.ErrorLevel: ErrorLevel,
		logrus.WarnLevel:  WarnLevel,
		logrus.InfoLevel:  InfoLevel,
		logrus.DebugLevel: DebugLevel,
	}
)

type LogrusWrapper struct {
	*logrus.Logger
}

func (l *LogrusWrapper) Level() LogLevel {
	return logrusLevelToLevel[l.Logger.Level]
}

func (l *LogrusWrapper) WithFields(fields LogFields) Logger {
	return &logrusEntryWrapper{Entry: l.Logger.WithFields(logrus.Fields(fields))}
}

type logrusEntryWrapper struct {
	*logrus.Entry
}

func (l logrusEntryWrapper) Level() LogLevel {
	return logrusLevelToLevel[l.Logger.Level]
}

func (l logrusEntryWrapper) WithFields(fields LogFields) Logger {
	return &logrusEntryWrapper{Entry: l.Logger.WithFields(logrus.Fields(fields))}
}

func (l *LogrusWrapper) SetLevel(level LogLevel) Logger {
	l.Logger.SetLevel(levelToLogrusLevel[level])
	return l
}
