package colorful

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/cocktail828/go-tools/xlog"
	"github.com/fatih/color"
)

type Logger struct {
	*log.Logger
	level                        xlog.Level
	debu, info, warn, erro, fata *color.Color
}

// an wrapper of *log.Logger with colorful output
func NewColorful(out io.Writer, prefix string, flag int) *Logger {
	return NewColorfulLog(log.New(out, prefix, flag))
}

func NewColorfulLog(l *log.Logger) *Logger {
	return &Logger{
		Logger: l,
		debu:   color.New().Add(color.Italic, color.Bold),
		info:   color.New(),
		warn:   color.New(color.FgYellow),
		erro:   color.New(color.FgRed),
		fata:   color.New(color.FgRed, color.Bold),
	}
}

func (l *Logger) WithColor(lv xlog.Level, c *color.Color) {
	switch lv {
	case xlog.LevelDebug:
		l.debu = c
	case xlog.LevelInfo:
		l.info = c
	case xlog.LevelWarn:
		l.warn = c
	case xlog.LevelError:
		l.erro = c
	case xlog.LevelFatal:
		l.fata = c
	}
}

func (l *Logger) iterate(f func(c *color.Color), levels ...xlog.Level) {
	for _, lv := range levels {
		switch lv {
		case xlog.LevelDebug:
			f(l.debu)
		case xlog.LevelInfo:
			f(l.info)
		case xlog.LevelWarn:
			f(l.warn)
		case xlog.LevelError:
			f(l.erro)
		case xlog.LevelFatal:
			f(l.fata)
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

func (l *Logger) log(depth int, lv xlog.Level, stringer func() string) {
	if lv >= l.level {
		l.Logger.Output(depth, stringer())
	}
}

func (l *Logger) Print(v ...any) {
	l.log(3, xlog.LevelFatal, func() string { return fmt.Sprint(v...) })
}

func (l *Logger) Println(v ...any) {
	l.log(3, xlog.LevelFatal, func() string { return fmt.Sprintln(v...) })
}

func (l *Logger) Printf(format string, v ...any) {
	l.log(3, xlog.LevelFatal, func() string { return fmt.Sprintf(format, v...) })
}

func (l *Logger) Debug(v ...any) {
	l.log(3, xlog.LevelDebug, func() string { return l.debu.Sprint(v...) })
}

func (l *Logger) Debugln(v ...any) {
	l.log(3, xlog.LevelDebug, func() string { return l.debu.Sprintln(v...) })
}

func (l *Logger) Debugf(format string, v ...any) {
	l.log(3, xlog.LevelDebug, func() string { return l.debu.Sprintf(format, v...) })
}

func (l *Logger) Info(v ...any) {
	l.log(3, xlog.LevelInfo, func() string { return l.info.Sprint(v...) })
}

func (l *Logger) Infoln(v ...any) {
	l.log(3, xlog.LevelInfo, func() string { return l.info.Sprintln(v...) })
}

func (l *Logger) Infof(format string, v ...any) {
	l.log(3, xlog.LevelInfo, func() string { return l.info.Sprintf(format, v...) })
}

func (l *Logger) Warn(v ...any) {
	l.log(3, xlog.LevelWarn, func() string { return l.warn.Sprint(v...) })
}

func (l *Logger) Warnln(v ...any) {
	l.log(3, xlog.LevelWarn, func() string { return l.warn.Sprintln(v...) })
}

func (l *Logger) Warnf(format string, v ...any) {
	l.log(3, xlog.LevelWarn, func() string { return l.warn.Sprintf(format, v...) })
}

func (l *Logger) Error(v ...any) {
	l.log(3, xlog.LevelError, func() string { return l.erro.Sprint(v...) })
}

func (l *Logger) Errorln(v ...any) {
	l.log(3, xlog.LevelError, func() string { return l.erro.Sprintln(v...) })
}

func (l *Logger) Errorf(format string, v ...any) {
	l.log(3, xlog.LevelError, func() string { return l.erro.Sprintf(format, v...) })
}

func (l *Logger) Fatal(v ...any) {
	l.log(3, xlog.LevelFatal, func() string { return l.fata.Sprint(v...) })
	os.Exit(1)
}

func (l *Logger) Fatalln(v ...any) {
	l.log(3, xlog.LevelFatal, func() string { return l.fata.Sprintln(v...) })
	os.Exit(1)
}

func (l *Logger) Fatalf(format string, v ...any) {
	l.log(3, xlog.LevelFatal, func() string { return l.fata.Sprintf(format, v...) })
	os.Exit(1)
}
