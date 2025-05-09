package colorful

import (
	"fmt"
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

func SetLevel(lv xlog.Level) {
	std.SetLevel(lv)
}

func Print(v ...any) {
	std.log(3, xlog.LevelFatal, func() string { return fmt.Sprint(v...) })
}

func Println(v ...any) {
	std.log(3, xlog.LevelFatal, func() string { return fmt.Sprintln(v...) })
}

func Printf(format string, v ...any) {
	std.log(3, xlog.LevelFatal, func() string { return fmt.Sprintf(format, v...) })
}

func Debug(v ...any) {
	std.log(3, xlog.LevelDebug, func() string { return std.debu.Sprint(v...) })
}

func Debugln(v ...any) {
	std.log(3, xlog.LevelDebug, func() string { return std.debu.Sprintln(v...) })
}

func Debugf(format string, v ...any) {
	std.log(3, xlog.LevelDebug, func() string { return std.debu.Sprintf(format, v...) })
}

func Info(v ...any) {
	std.log(3, xlog.LevelInfo, func() string { return std.info.Sprint(v...) })
}

func Infoln(v ...any) {
	std.log(3, xlog.LevelInfo, func() string { return std.info.Sprintln(v...) })
}

func Infof(format string, v ...any) {
	std.log(3, xlog.LevelInfo, func() string { return std.info.Sprintf(format, v...) })
}

func Warn(v ...any) {
	std.log(3, xlog.LevelWarn, func() string { return std.warn.Sprint(v...) })
}

func Warnln(v ...any) {
	std.log(3, xlog.LevelWarn, func() string { return std.warn.Sprintln(v...) })
}

func Warnf(format string, v ...any) {
	std.log(3, xlog.LevelWarn, func() string { return std.warn.Sprintf(format, v...) })
}

func Error(v ...any) {
	std.log(3, xlog.LevelError, func() string { return std.erro.Sprint(v...) })
}

func Errorln(v ...any) {
	std.log(3, xlog.LevelError, func() string { return std.erro.Sprintln(v...) })
}

func Errorf(format string, v ...any) {
	std.log(3, xlog.LevelError, func() string { return std.erro.Sprintf(format, v...) })
}

func Fatal(v ...any) {
	std.log(3, xlog.LevelFatal, func() string { return std.fata.Sprint(v...) })
	os.Exit(1)
}

func Fatalln(v ...any) {
	std.log(3, xlog.LevelFatal, func() string { return std.fata.Sprintln(v...) })
	os.Exit(1)
}

func Fatalf(format string, v ...any) {
	std.log(3, xlog.LevelFatal, func() string { return std.fata.Sprintf(format, v...) })
	os.Exit(1)
}
