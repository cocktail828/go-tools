package xlog

import (
	"os"
	"testing"
)

func TestLog(t *testing.T) {
	l := NewJSONHandler(
		os.Stderr, HandlerOptions{
			AddSource: false,
			Level:     LevelDebug,
		},
	)

	lg := l.WithGroup("aaaa")
	lg.Debugln("Debug", "key1", "val1", "key2", "val2")
	lg.Infoln("Info", "key1", "val1", "key2", "val2")
	lg.Warnln("Warn", "key1", "val1", "key2", "val2")
	lg.Errorln("Error", "key1", "val1", "key2", "val2")

	lg.Debugf("Debugf", "xxxx %v", "123")
	lg.Infof("Infof", "xxxx %v", "123")
	lg.Warnf("Warnf", "xxxx %v", "123")
	lg.Errorf("Errorf", "xxxx %v", "123")
}
