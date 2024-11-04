package main

import (
	"log"

	"github.com/cocktail828/go-tools/xlog/colorful"
)

func dump(l *colorful.Logger) {
	l.Debug("Debug", "key1", "val1", "key2", "val2")
	l.Debugln("Debug", "key1", "val1", "key2", "val2")
	l.Debugf("Debugf xxxx %v", "123")

	l.Info("Info", "key1", "val1", "key2", "val2")
	l.Infoln("Info", "key1", "val1", "key2", "val2")
	l.Infof("Infof xxxx %v", "123")

	l.Warn("Warn", "key1", "val1", "key2", "val2")
	l.Warnln("Info", "key1", "val1", "key2", "val2")
	l.Warnf("Warnf xxxx %v", "123")

	l.Error("Error", "key1", "val1", "key2", "val2")
	l.Errorln("Error", "key1", "val1", "key2", "val2")
	l.Errorf("Errorf xxxx %v", "123")

	// l.Fatal("Fatal", "key1", "val1", "key2", "val2")
	// l.Fatalln("Fatal", "key1", "val1", "key2", "val2")
	// l.Fatalf("Fatalf xxxx %v", "123")
}

func main() {
	log.SetFlags(log.Flags() | log.Llongfile)
	l0 := colorful.NewColorfulLog(log.Default())
	l1 := colorful.Default()
	// dump(l)
	// dump(l)
	// l.DisableColor()
	// dump(l)
	// l.EnableColor()
	// dump(l)

	l0.Warn("xxxxxxxxxxx")
	l1.Warn("xxxxxxxxxxx")
	colorful.Warn("xxxxxxxxxxx")
}
