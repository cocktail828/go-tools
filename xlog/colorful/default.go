package colorful

import (
	"log"
	"os"

	"github.com/cocktail828/go-tools/xlog"
	"github.com/fatih/color"
)

var (
	std = NewColorfulLog(log.Default())
)

func Default() *Logger { return std }

func SetColor(lv xlog.Level, c *color.Color) { std.SetColor(lv, c) }

// disable all color if no level is passed
func DisableColor(levels ...xlog.Level) { std.DisableColor() }

// enable all color if no level is passed
func EnableColor(levels ...xlog.Level) { std.EnableColor(levels...) }

func SetLevel(lv xlog.Level) { std.SetLevel(lv) }
func GetLevel() xlog.Level   { return std.GetLevel() }

func Print(v ...any)                 { std.log(3, std.print, v...) }
func Println(v ...any)               { std.logln(3, std.print, v...) }
func Printf(format string, v ...any) { std.logf(3, std.print, format, v...) }

func Debug(v ...any)                 { std.log(3, std.debu, v...) }
func Debugln(v ...any)               { std.logln(3, std.debu, v...) }
func Debugf(format string, v ...any) { std.logf(3, std.debu, format, v...) }

func Info(v ...any)                 { std.log(3, std.info, v...) }
func Infoln(v ...any)               { std.logln(3, std.info, v...) }
func Infof(format string, v ...any) { std.logf(3, std.info, format, v...) }

func Warn(v ...any)                 { std.log(3, std.warn, v...) }
func Warnln(v ...any)               { std.logln(3, std.warn, v...) }
func Warnf(format string, v ...any) { std.logf(3, std.warn, format, v...) }

func Error(v ...any)                 { std.log(3, std.erro, v...) }
func Errorln(v ...any)               { std.logln(3, std.erro, v...) }
func Errorf(format string, v ...any) { std.logf(3, std.erro, format, v...) }

func Fatal(v ...any)                 { std.log(3, std.fata, v...); os.Exit(1) }
func Fatalln(v ...any)               { std.logln(3, std.fata, v...); os.Exit(1) }
func Fatalf(format string, v ...any) { std.logf(3, std.fata, format, v...); os.Exit(1) }
