package colorful

import (
	"io"
	"log"
	"os"

	"github.com/cocktail828/go-tools/xlog"
	"github.com/fatih/color"
)

type lvprinter interface {
	Level() xlog.Level
	Sprint(v ...any) string
	Sprintln(v ...any) string
	Sprintf(format string, v ...any) string
}

type lvcolor struct {
	lv xlog.Level
	*color.Color
}

func (lc lvcolor) Level() xlog.Level { return lc.lv }

type Logger struct {
	*log.Logger
	level                               xlog.Level
	print, debu, info, warn, erro, fata *lvcolor
}

// an wrapper of *log.Logger with colorful output
func NewColorful(out io.Writer, prefix string, flag int) *Logger {
	return NewColorfulLog(log.New(out, prefix, flag))
}

func NewColorfulLog(l *log.Logger) *Logger {
	return &Logger{
		Logger: l,
		print:  &lvcolor{xlog.LevelFatal, color.New()}, // The printer will definitely be able to print out the log with high log level.
		debu:   &lvcolor{xlog.LevelDebug, color.New().Add(color.Italic, color.Bold)},
		info:   &lvcolor{xlog.LevelInfo, color.New()},
		warn:   &lvcolor{xlog.LevelWarn, color.New(color.FgYellow)},
		erro:   &lvcolor{xlog.LevelError, color.New(color.FgRed)},
		fata:   &lvcolor{xlog.LevelFatal, color.New(color.FgRed, color.Bold)},
	}
}

func (l *Logger) SetColor(lv xlog.Level, c *color.Color) {
	colors := []*lvcolor{l.debu, l.info, l.warn, l.erro, l.fata}
	for _, p := range colors {
		if p.lv == lv && c != nil {
			p.Color = c
		}
	}
}

func (l *Logger) iterate(f func(c *color.Color), levels ...xlog.Level) {
	colors := []*lvcolor{l.debu, l.info, l.warn, l.erro, l.fata}
	for _, lv := range levels {
		for _, p := range colors {
			if p.lv == lv && f != nil {
				f(p.Color)
			}
		}
	}
}

// disable all color if no level is passed
func (l *Logger) DisableColor(levels ...xlog.Level) {
	l.iterate(func(c *color.Color) { c.DisableColor() }, levels...)
}

// enable all color if no level is passed
func (l *Logger) EnableColor(levels ...xlog.Level) {
	l.iterate(func(c *color.Color) { c.EnableColor() }, levels...)
}

func (l *Logger) SetLevel(lv xlog.Level) {
	l.level = lv
}

func (l *Logger) log(depth int, printer lvprinter, v ...any) {
	if printer.Level() >= l.level {
		l.Logger.Output(depth, printer.Sprint(v...))
	}
}

func (l *Logger) logln(depth int, printer lvprinter, v ...any) {
	if printer.Level() >= l.level {
		l.Logger.Output(depth, printer.Sprintln(v...))
	}
}

func (l *Logger) logf(depth int, printer lvprinter, format string, v ...any) {
	if printer.Level() >= l.level {
		l.Logger.Output(depth, printer.Sprintf(format, v...))
	}
}

func (l *Logger) Print(v ...any)                 { l.log(3, l.print, v...) }
func (l *Logger) Println(v ...any)               { l.logln(3, l.print, v...) }
func (l *Logger) Printf(format string, v ...any) { l.logf(3, l.print, format, v...) }

func (l *Logger) Debug(v ...any)                 { l.log(3, l.debu, v...) }
func (l *Logger) Debugln(v ...any)               { l.logln(3, l.debu, v...) }
func (l *Logger) Debugf(format string, v ...any) { l.logf(3, l.debu, format, v...) }

func (l *Logger) Info(v ...any)                 { l.log(3, l.info, v...) }
func (l *Logger) Infoln(v ...any)               { l.logln(3, l.info, v...) }
func (l *Logger) Infof(format string, v ...any) { l.logf(3, l.info, format, v...) }

func (l *Logger) Warn(v ...any)                 { l.log(3, l.warn, v...) }
func (l *Logger) Warnln(v ...any)               { l.logln(3, l.warn, v...) }
func (l *Logger) Warnf(format string, v ...any) { l.logf(3, l.warn, format, v...) }

func (l *Logger) Error(v ...any)                 { l.log(3, l.erro, v...) }
func (l *Logger) Errorln(v ...any)               { l.logln(3, l.erro, v...) }
func (l *Logger) Errorf(format string, v ...any) { l.logf(3, l.erro, format, v...) }

func (l *Logger) Fatal(v ...any)                 { l.log(3, l.fata, v...); os.Exit(1) }
func (l *Logger) Fatalln(v ...any)               { l.logln(3, l.fata, v...); os.Exit(1) }
func (l *Logger) Fatalf(format string, v ...any) { l.logf(3, l.fata, format, v...); os.Exit(1) }
