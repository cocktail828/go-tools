package xlog

import (
	"log/slog"
)

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

// HandlerOptions are options for a [TextHandler] or [JSONHandler].
// A zero HandlerOptions consists entirely of default values.
type HandlerOptions struct {
	// AddSource causes the handler to compute the source code position
	// of the log statement and add a SourceKey attribute to the output.
	AddSource bool

	// Level reports the minimum record level that will be logged.
	// The handler discards records with lower levels.
	// If Level is nil, the handler assumes LevelInfo.
	// The handler calls Level.Level for each record processed;
	// to adjust the minimum level dynamically, use a LevelVar.
	Level Level
}

type Logger interface {
	Debugf(format string, args ...any)
	Debugln(msg string, args ...any)
	Errorf(format string, args ...any)
	Errorln(msg string, args ...any)
	Fatalf(format string, args ...any)
	Fatalln(msg string, args ...any)
	Infof(format string, args ...any)
	Infoln(msg string, args ...any)
	Printf(format string, args ...any)
	Println(msg string, args ...any)
	Warnf(format string, args ...any)
	Warnln(msg string, args ...any)
	With(args ...any) Logger
	WithGroup(name string) Logger
}
