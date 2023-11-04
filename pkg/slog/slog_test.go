package slog_test

import (
	"io"
	"os"
	"testing"

	"golang.org/x/exp/slog"
	"gopkg.in/natefinch/lumberjack.v2"
)

func TestSlog(t *testing.T) {
	var lvl slog.LevelVar
	lvl.Set(slog.LevelDebug)

	l := slog.New(slog.NewJSONHandler(io.MultiWriter(os.Stderr, &lumberjack.Logger{
		Filename:   "error.log",
		MaxSize:    100,
		MaxBackups: 1,
		MaxAge:     1,
		Compress:   false,
		LocalTime:  false,
	}), &slog.HandlerOptions{Level: &lvl}).WithAttrs([]slog.Attr{slog.String("a", "b")}))
	l = l.With("a1", "b1")
	l.Info("finished", "key", "value")
	l.Info("finishedxxx", "key", "value")
}
