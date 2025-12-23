package xlog

type Printer interface {
	Printf(format string, v ...any)
}

type NoopPrinter struct{}

func (p NoopPrinter) Printf(format string, v ...any) {}

// A Level is the importance or severity of a log event.
// The higher the level, the more important or severe the event.
type Level int

const (
	LevelDebug Level = iota
	LevelInfo  Level = iota
	LevelWarn  Level = iota
	LevelError Level = iota
	LevelFatal Level = iota
)

var AllLevels = []Level{LevelDebug, LevelInfo, LevelWarn, LevelError, LevelFatal}
