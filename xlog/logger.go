package xlog

type Printer interface {
	Printf(format string, v ...any)
}

type NopPrinter struct{}

func (p NopPrinter) Printf(format string, v ...any) {}

// A Level is the importance or severity of a log event.
// The higher the level, the more important or severe the event.
type Level int

const (
	LevelDebug Level = iota - 1
	LevelInfo  Level = iota - 1
	LevelWarn  Level = iota - 1
	LevelError Level = iota - 1
	LevelFatal Level = iota - 1
)

var AllLevels = []Level{LevelDebug, LevelInfo, LevelWarn, LevelError, LevelFatal}
