package colorful

import (
	"fmt"
	"log"
	"testing"
)

func dump(prefix string, l *Logger) {
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

func TestColorful(t *testing.T) {
	func() {
		fmt.Println("=============== default ==============")
		log.SetFlags(log.Flags() | log.Llongfile)
		dump("default", Default())
	}()

	func() {
		fmt.Println("=============== new ==============")
		log.SetFlags(log.Flags() | log.Llongfile)
		dump("new", NewColorfulLog(log.Default()))
	}()
}
