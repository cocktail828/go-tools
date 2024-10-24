package colorful

import (
	"io"
	"log"
	"os"

	"github.com/cocktail828/go-tools/xlog"
	"github.com/fatih/color"
)

type printer struct {
	*color.Color
	level xlog.Level
}

func (p printer) Level() xlog.Level { return p.level }

type Logger struct {
	*log.Logger
	debu, info, warn, erro, fata printer
}

// an wrapper of *log.Logger with colorful output
func NewColorful(out io.Writer, prefix string, flag int) *Logger {
	cl := &Logger{Logger: log.New(out, prefix, flag)}
	cl.init()
	return cl
}

func NewColorfulLog(l *log.Logger) *Logger {
	cl := &Logger{Logger: l}
	cl.init()
	return cl
}

func (l *Logger) init() {
	l.debu = printer{color.New().Add(color.Underline), xlog.LevelDebug}
	l.info = printer{color.New(), xlog.LevelInfo}
	l.warn = printer{color.New(color.FgYellow), xlog.LevelWarn}
	l.erro = printer{color.New(color.FgRed), xlog.LevelError}
	l.fata = printer{color.New(color.FgRed, color.Bold), xlog.LevelFatal}
}

// disable all color if no level is passed
func (l *Logger) DisableColor(levels ...xlog.Level) {
	if len(levels) == 0 {
		levels = xlog.AllLevels
	}
	pr := []printer{l.debu, l.info, l.warn, l.erro, l.fata}
	for _, p := range pr {
		for _, lv := range levels {
			if p.Level() == lv {
				p.DisableColor()
			}
		}
	}
}

// enable all color if no level is passed
func (l *Logger) EnableColor(levels ...xlog.Level) {
	if len(levels) == 0 {
		levels = xlog.AllLevels
	}
	pr := []printer{l.debu, l.info, l.warn, l.erro, l.fata}
	for _, p := range pr {
		for _, lv := range levels {
			if p.Level() == lv {
				p.EnableColor()
			}
		}
	}
}

func (l *Logger) log(msg string) {
	l.Logger.Output(3, msg)
}

func (l *Logger) Debug(v ...any) {
	l.log(l.debu.Sprint(v...))
}

func (l *Logger) Debugln(v ...any) {
	l.log(l.debu.Sprintln(v...))
}

func (l *Logger) Debugf(format string, v ...any) {
	l.log(l.debu.Sprintf(format, v...))
}

func (l *Logger) Info(v ...any) {
	l.log(l.info.Sprint(v...))
}

func (l *Logger) Infoln(v ...any) {
	l.log(l.info.Sprintln(v...))
}

func (l *Logger) Infof(format string, v ...any) {
	l.log(l.info.Sprintf(format, v...))
}

func (l *Logger) Warn(v ...any) {
	l.log(l.warn.Sprint(v...))
}

func (l *Logger) Warnln(v ...any) {
	l.log(l.warn.Sprintln(v...))
}

func (l *Logger) Warnf(format string, v ...any) {
	l.log(l.warn.Sprintf(format, v...))
}

func (l *Logger) Error(v ...any) {
	l.log(l.erro.Sprint(v...))
}

func (l *Logger) Errorln(v ...any) {
	l.log(l.erro.Sprintln(v...))
}

func (l *Logger) Errorf(format string, v ...any) {
	l.log(l.erro.Sprintf(format, v...))
}

func (l *Logger) Fatal(v ...any) {
	l.log(l.fata.Sprint(v...))
	os.Exit(1)
}

func (l *Logger) Fatalln(v ...any) {
	l.log(l.fata.Sprintln(v...))
	os.Exit(1)
}

func (l *Logger) Fatalf(format string, v ...any) {
	l.log(l.fata.Sprintf(format, v...))
	os.Exit(1)
}
