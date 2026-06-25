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
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

var AllLevels = []Level{LevelDebug, LevelInfo, LevelWarn, LevelError, LevelFatal}
