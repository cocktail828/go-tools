package main

import (
	"fmt"
	"log"

	"github.com/cocktail828/go-tools/xlog"
	"github.com/cocktail828/go-tools/xlog/colorful"
)

func dump(prefix string, l *colorful.Logger) {
	l.Debug(prefix+".Debug", "key1", "val1", "key2", "val2")
	l.Debugln(prefix+".Debugln", "key1", "val1", "key2", "val2")
	l.Debugf(prefix+".Debugf xxxx %v", "123")

	l.Info(prefix+".Info", "key1", "val1", "key2", "val2")
	l.Infoln(prefix+".Infoln", "key1", "val1", "key2", "val2")
	l.Infof(prefix+".Infof xxxx %v", "123")

	l.Warn(prefix+".Warn", "key1", "val1", "key2", "val2")
	l.Warnln(prefix+".Infoln", "key1", "val1", "key2", "val2")
	l.Warnf(prefix+".Warnf xxxx %v", "123")

	l.Error(prefix+".Error", "key1", "val1", "key2", "val2")
	l.Errorln(prefix+".Errorln", "key1", "val1", "key2", "val2")
	l.Errorf(prefix+".Errorf xxxx %v", "123")
}

func main() {
	log.SetFlags(log.Flags() | log.Llongfile)

	func() {
		fmt.Println("=============== default ==============")
		dump("default", colorful.Default())
	}()

	func() {
		fmt.Println("=============== new ==============")
		dump("new", colorful.NewColorfulLog(log.Default()))
	}()

	func() {
		fmt.Println("=============== level ==============")
		l := colorful.NewColorfulLog(log.Default())
		l.SetLevel(xlog.LevelDebug)
		dump("level", l)
	}()
}
