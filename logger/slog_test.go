package logger_test

import (
	"testing"

	"github.com/cocktail828/go-tools/logger"
)

func TestSlog(t *testing.T) {
	l := logger.NewLoggerWithLumberjack(logger.Config{
		Level:      "error",
		Filename:   "/log/server/error.log",
		MaxSize:    100,
		MaxCount: 1,
		MaxAge:     1,
		Compress:   false,
	})
	l = l.With("a1", "b1").WithGroup("xxx")
	l.Infow("slog.finished", "key", "value")
	l.Errorw("slog.finishedxxx", "key", "value")
	l.Errorf("slog.finishedxxx %v", "key")
}
