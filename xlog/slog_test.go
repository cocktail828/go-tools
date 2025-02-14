package xlog_test

import (
	"os"
	"testing"

	"github.com/cocktail828/go-tools/xlog"
)

func TestLog(t *testing.T) {
	l := xlog.NewJSONHandler(
		os.Stderr, xlog.HandlerOptions{
			AddSource: false,
			Level:     xlog.LevelWarn,
		},
	)

	lg := l.WithGroup("aaaa")
	lg.Debugln("xlog.Debug", "key1", "val1", "key2", "val2")
	lg.Debugf("xlog.Debugf xxxx %v", "123")

	lg.Infoln("xlog.Info", "key1", "val1", "key2", "val2")
	lg.Infof("xlog.Infof xxxx %v", "123")

	lg.Warnln("xlog.Warn", "key1", "val1", "key2", "val2")
	lg.Warnf("xlog.Warnf xxxx %v", "123")

	lg.Errorln("xlog.Error", "key1", "val1", "key2", "val2")
	lg.Errorf("xlog.Errorf xxxx %v", "123")
}
