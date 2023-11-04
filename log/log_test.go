package log_test

import (
	"testing"

	"github.com/cocktail828/go-tools/log"
	"github.com/sirupsen/logrus"
)

func init() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.InfoLevel) // default
	logrus.AddHook(log.NewLumberjack(
		log.LogFile{
			Filename:   "error.log",
			MaxSize:    100,
			MaxBackups: 1,
			MaxAge:     1,
			Compress:   false,
			LocalTime:  false,
		},
		&logrus.TextFormatter{DisableColors: true},
		logrus.ErrorLevel,
	))
}

func TestXXX(t *testing.T) {
	logrus.Trace("Trace msg") // not be written
	logrus.Warn("Warn msg")   // written in general.log
	logrus.Error("Error msg") // written in error.log
}
