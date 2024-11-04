package colorful

import (
	"log"
	"os"

	"github.com/cocktail828/go-tools/xlog"
)

var (
	std = NewColorfulLog(log.Default())
)

func Default() *Logger { return std }

// disable all color if no level is passed
func DisableColor(levels ...xlog.Level) {
	std.DisableColor()
}

// enable all color if no level is passed
func EnableColor(levels ...xlog.Level) {
	std.EnableColor(levels...)
}

func Debug(v ...any) {
	std.log(3, std.debu.Sprint(v...))
}

func Debugln(v ...any) {
	std.log(3, std.debu.Sprintln(v...))
}

func Debugf(format string, v ...any) {
	std.log(3, std.debu.Sprintf(format, v...))
}

func Info(v ...any) {
	std.log(3, std.info.Sprint(v...))
}

func Infoln(v ...any) {
	std.log(3, std.info.Sprintln(v...))
}

func Infof(format string, v ...any) {
	std.log(3, std.info.Sprintf(format, v...))
}

func Warn(v ...any) {
	std.log(3, std.warn.Sprint(v...))
}

func Warnln(v ...any) {
	std.log(3, std.warn.Sprintln(v...))
}

func Warnf(format string, v ...any) {
	std.log(3, std.warn.Sprintf(format, v...))
}

func Error(v ...any) {
	std.log(3, std.erro.Sprint(v...))
}

func Errorln(v ...any) {
	std.log(3, std.erro.Sprintln(v...))
}

func Errorf(format string, v ...any) {
	std.log(3, std.erro.Sprintf(format, v...))
}

func Fatal(v ...any) {
	std.log(3, std.fata.Sprint(v...))
	os.Exit(1)
}

func Fatalln(v ...any) {
	std.log(3, std.fata.Sprintln(v...))
	os.Exit(1)
}

func Fatalf(format string, v ...any) {
	std.log(3, std.fata.Sprintf(format, v...))
	os.Exit(1)
}
