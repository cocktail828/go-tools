package main

import (
	"os"

	"github.com/cocktail828/go-tools/xlog"
	"github.com/cocktail828/go-tools/xlog/colorful"
)

func main() {
	logs := []func() *colorful.Logger{
		func() *colorful.Logger {
			l := colorful.Default()
			l.SetPrefix("[default] ")
			l.SetFlags(l.Flags() | colorful.Llongfile)
			return l
		},
		func() *colorful.Logger {
			l := colorful.NewColorful(os.Stderr, "", colorful.LstdFlags|colorful.Llongfile)
			l.SetPrefix("[new] ")
			return l
		},
	}

	for _, f := range logs {
		l := f()
		l.SetLevel(xlog.LevelDebug)

		l.Debug("Debug", "key1", "val1", "key2", "val2")
		l.Debugln("Debugln", "key1", "val1", "key2", "val2")
		l.Debugf("Debugf xxxx %v", "123")

		l.Info("Info", "key1", "val1", "key2", "val2")
		l.Infoln("Infoln", "key1", "val1", "key2", "val2")
		l.Infof("Infof xxxx %v", "123")

		l.Warn("Warn", "key1", "val1", "key2", "val2")
		l.Warnln("Infoln", "key1", "val1", "key2", "val2")
		l.Warnf("Warnf xxxx %v", "123")

		l.Error("Error", "key1", "val1", "key2", "val2")
		l.Errorln("Errorln", "key1", "val1", "key2", "val2")
		l.Errorf("Errorf xxxx %v", "123")
	}
}
