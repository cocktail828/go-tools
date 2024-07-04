package logger

import (
	"strings"

	"github.com/cocktail828/go-tools/pkg/lumberjack.v2"
	"github.com/cocktail828/go-tools/pkg/slog"
)

func NewLoggerWithLumberjack(cfg Config) Logger {
	var lvl slog.LevelVar
	lvl.Set(func() slog.Level {
		switch strings.ToLower(cfg.Level) {
		case "debug":
			return slog.LevelDebug
		case "info":
			return slog.LevelInfo
		case "warn":
			return slog.LevelWarn
		case "error":
			return slog.LevelError
		default:
			return slog.LevelError
		}
	}())

	return NewLoggerWithSlog(slog.New(slog.NewJSONHandler(
		&lumberjack.Logger{
			Filename:   cfg.Filename,
			MaxSize:    cfg.MaxSize,
			Async:      cfg.Async,
			BufSize:    cfg.BufSize,
			MaxAge:     cfg.MaxAge,
			MaxBackups: cfg.MaxCount,
			Compress:   cfg.Compress,
		}, &slog.HandlerOptions{
			AddSource: cfg.AddSource,
			Level:     &lvl,
		},
	)))
}
