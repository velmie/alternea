package bootstrap

import (
	"github.com/sirupsen/logrus"

	"github.com/velmie/alternea/app"
)

var logger app.Logger

func SetLogger(log app.Logger) {
	logger = log
}

func GetLogger() app.Logger {
	if logger != nil {
		return logger
	}
	logger = (&app.LogrusWrapper{Logger: logrus.New()}).SetLevel(app.InfoLevel)
	return logger
}
