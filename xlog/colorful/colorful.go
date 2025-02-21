package colorful

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/cocktail828/go-tools/xlog"
	"github.com/fatih/color"
)

type Color struct {
	*color.Color
	Level xlog.Level
}

type Logger struct {
	*log.Logger
	debu, info, warn, erro, fata Color
}

// an wrapper of *log.Logger with colorful output
func NewColorful(out io.Writer, prefix string, flag int) *Logger {
	return NewColorfulLog(log.New(out, prefix, flag))
}

func NewColorfulLog(l *log.Logger) *Logger {
	return &Logger{
		Logger: l,
		debu:   Color{color.New().Add(color.Italic, color.Bold), xlog.LevelDebug},
		info:   Color{color.New(), xlog.LevelInfo},
		warn:   Color{color.New(color.FgYellow), xlog.LevelWarn},
		erro:   Color{color.New(color.FgRed), xlog.LevelError},
		fata:   Color{color.New(color.FgRed, color.Bold), xlog.LevelFatal},
	}
}

func (l *Logger) WithColor(c Color) {
	switch c.Level {
	case xlog.LevelDebug:
		l.debu.Color = c.Color
	case xlog.LevelInfo:
		l.info.Color = c.Color
	case xlog.LevelWarn:
		l.warn.Color = c.Color
	case xlog.LevelError:
		l.erro.Color = c.Color
	case xlog.LevelFatal:
		l.fata.Color = c.Color
	}
}

// disable all color if no level is passed
func (l *Logger) DisableColor(levels ...xlog.Level) {
	pr := []Color{l.debu, l.info, l.warn, l.erro, l.fata}
	for _, p := range pr {
		for _, lv := range levels {
			if p.Level == lv {
				p.DisableColor()
			}
		}
	}
}

// enable all color if no level is passed
func (l *Logger) EnableColor(levels ...xlog.Level) {
	pr := []Color{l.debu, l.info, l.warn, l.erro, l.fata}
	for _, p := range pr {
		for _, lv := range levels {
			if p.Level == lv {
				p.EnableColor()
			}
		}
	}
}

func (l *Logger) log(depth int, msg string) {
	l.Logger.Output(depth, msg)
}

func (l *Logger) Print(v ...any) {
	l.log(3, fmt.Sprint(v...))
}

func (l *Logger) Println(v ...any) {
	l.log(3, fmt.Sprintln(v...))
}

func (l *Logger) Printf(format string, v ...any) {
	l.log(3, fmt.Sprintf(format, v...))
}

func (l *Logger) Debug(v ...any) {
	l.log(3, l.debu.Sprint(v...))
}

func (l *Logger) Debugln(v ...any) {
	l.log(3, l.debu.Sprintln(v...))
}

func (l *Logger) Debugf(format string, v ...any) {
	l.log(3, l.debu.Sprintf(format, v...))
}

func (l *Logger) Info(v ...any) {
	l.log(3, l.info.Sprint(v...))
}

func (l *Logger) Infoln(v ...any) {
	l.log(3, l.info.Sprintln(v...))
}

func (l *Logger) Infof(format string, v ...any) {
	l.log(3, l.info.Sprintf(format, v...))
}

func (l *Logger) Warn(v ...any) {
	l.log(3, l.warn.Sprint(v...))
}

func (l *Logger) Warnln(v ...any) {
	l.log(3, l.warn.Sprintln(v...))
}

func (l *Logger) Warnf(format string, v ...any) {
	l.log(3, l.warn.Sprintf(format, v...))
}

func (l *Logger) Error(v ...any) {
	l.log(3, l.erro.Sprint(v...))
}

func (l *Logger) Errorln(v ...any) {
	l.log(3, l.erro.Sprintln(v...))
}

func (l *Logger) Errorf(format string, v ...any) {
	l.log(3, l.erro.Sprintf(format, v...))
}

func (l *Logger) Fatal(v ...any) {
	l.log(3, l.fata.Sprint(v...))
	os.Exit(1)
}

func (l *Logger) Fatalln(v ...any) {
	l.log(3, l.fata.Sprintln(v...))
	os.Exit(1)
}

func (l *Logger) Fatalf(format string, v ...any) {
	l.log(3, l.fata.Sprintf(format, v...))
	os.Exit(1)
}
