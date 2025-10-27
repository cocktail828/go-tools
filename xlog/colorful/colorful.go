package colorful

import (
	"io"
	"log"
	"os"
	"sync/atomic"

	"github.com/cocktail828/go-tools/xlog"
	"github.com/fatih/color"
	"golang.org/x/time/rate"
)

type Flag int

const (
	Ldate         Flag = log.Ldate         // the date in the local time zone: 2009/01/23
	Ltime         Flag = log.Ltime         // the time in the local time zone: 01:23:23
	Lmicroseconds Flag = log.Lmicroseconds // microsecond resolution: 01:23:23.123123.  assumes Ltime.
	Llongfile     Flag = log.Llongfile     // full file name and line number: /a/b/c/d.go:23
	Lshortfile    Flag = log.Lshortfile    // final file name element and line number: d.go:23. overrides Llongfile
	LUTC          Flag = log.LUTC          // if Ldate or Ltime is set, use UTC rather than the local time zone
	Lmsgprefix    Flag = log.Lmsgprefix    // move the "prefix" from the beginning of the line to before the message
	LstdFlags     Flag = log.LstdFlags     // initial values for the standard logger
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
	stdlog                              *log.Logger
	level                               xlog.Level
	print, debu, info, warn, erro, fata *lvcolor
	panic                               *color.Color
}

// an wrapper of *log.Logger with colorful output
func NewColorful(out io.Writer, prefix string, flag Flag) *Logger {
	return newColorful(log.New(out, prefix, int(flag)))
}

func newColorful(l *log.Logger) *Logger {
	return &Logger{
		stdlog: l,
		print:  &lvcolor{xlog.LevelFatal, color.New()}, // The printer will definitely be able to print out the log with high log level.
		debu:   &lvcolor{xlog.LevelDebug, color.New().Add(color.Italic, color.FgGreen)},
		info:   &lvcolor{xlog.LevelInfo, color.New()},
		warn:   &lvcolor{xlog.LevelWarn, color.New(color.FgYellow)},
		erro:   &lvcolor{xlog.LevelError, color.New(color.FgRed)},
		fata:   &lvcolor{xlog.LevelFatal, color.New(color.FgRed, color.Bold)},
		panic:  color.New(color.FgRed, color.Bold),
	}
}

func (l *Logger) Flags() Flag                          { return Flag(l.stdlog.Flags()) }
func (l *Logger) Output(calldepth int, s string) error { return l.stdlog.Output(calldepth, s) }
func (l *Logger) Prefix() string                       { return l.stdlog.Prefix() }
func (l *Logger) SetFlags(flag Flag)                   { l.stdlog.SetFlags(int(flag)) }
func (l *Logger) SetOutput(w io.Writer)                { l.stdlog.SetOutput(w) }
func (l *Logger) SetPrefix(prefix string)              { l.stdlog.SetPrefix(prefix) }
func (l *Logger) Writer() io.Writer                    { return l.stdlog.Writer() }

func (l *Logger) SetColor(lv xlog.Level, c *color.Color) {
	colors := []*lvcolor{l.debu, l.info, l.warn, l.erro, l.fata}
	for _, p := range colors {
		if p.lv == lv && c != nil {
			p.Color = c
		}
	}
}

func (l *Logger) iterate(f func(c *color.Color), levels ...xlog.Level) {
	if len(levels) == 0 {
		levels = xlog.AllLevels
	}

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

func (l *Logger) SetLevel(lv xlog.Level) { l.level = lv }
func (l *Logger) GetLevel() xlog.Level   { return l.level }

func (l *Logger) log(depth int, printer lvprinter, v ...any) {
	if printer.Level() >= l.level {
		l.stdlog.Output(depth, printer.Sprint(v...))
	}
}

func (l *Logger) logln(depth int, printer lvprinter, v ...any) {
	if printer.Level() >= l.level {
		l.stdlog.Output(depth, printer.Sprintln(v...))
	}
}

func (l *Logger) logf(depth int, printer lvprinter, format string, v ...any) {
	if printer.Level() >= l.level {
		l.stdlog.Output(depth, printer.Sprintf(format, v...))
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

func (l *Logger) Panic(v ...any)                 { panic(l.panic.Sprint(v...)) }
func (l *Logger) Panicln(v ...any)               { panic(l.panic.Sprintln(v...)) }
func (l *Logger) Panicf(format string, v ...any) { panic(l.panic.Sprintf(format, v...)) }

// Limited returns a limitedColorful logger.
// The logger will only print log if the limiter allows.
// It's useful when too much same log happens, and we want to limit the log output.
func (l *Logger) Limited(limiter *rate.Limiter) *limitedColorful {
	return &limitedColorful{
		Logger:  l,
		limiter: limiter,
		supressMap: map[xlog.Level]*atomic.Uint32{
			xlog.LevelDebug: {},
			xlog.LevelInfo:  {},
			xlog.LevelWarn:  {},
			xlog.LevelError: {},
		},
	}
}
