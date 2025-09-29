package xlog

import "log/slog"

type Printer interface {
	Printf(format string, v ...any)
}

type NoopPrinter struct{}

func (p NoopPrinter) Printf(format string, v ...any) {}

// A Level is the importance or severity of a log event.
// The higher the level, the more important or severe the event.
type Level int

const (
	LevelDebug Level = Level(slog.LevelDebug)
	LevelInfo  Level = Level(slog.LevelInfo)
	LevelWarn  Level = Level(slog.LevelWarn)
	LevelError Level = Level(slog.LevelError)
	LevelFatal Level = Level(slog.LevelError + 4)
)

func (l Level) Level() slog.Level { return slog.Level(l) }

var AllLevels = []Level{LevelDebug, LevelInfo, LevelWarn, LevelError, LevelFatal}
