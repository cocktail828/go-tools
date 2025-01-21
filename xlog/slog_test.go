package xlog_test

import (
	"os"
	"testing"

	"github.com/cocktail828/go-tools/xlog"
)

func TestLog(t *testing.T) {
	cases := []struct {
		name string
		lg   xlog.Logger
	}{
		{
			"New-json",
			xlog.NewJSONHandler(
				os.Stderr, xlog.HandlerOptions{
					AddSource: true,
					Level:     xlog.LevelDebug,
				},
			),
		},
		{
			"New-text",
			xlog.NewTextHandler(
				os.Stderr, xlog.HandlerOptions{
					AddSource: true,
					Level:     xlog.LevelDebug,
				},
			),
		},
	}
	for _, l := range cases {
		lg := l.lg.WithGroup("aaaa")
		lg.Debugln(l.name+".Debug", "key1", "val1", "key2", "val2")
		lg.Debugf(l.name+".Debugf xxxx %v", "123")

		lg.Infoln(l.name+".Info", "key1", "val1", "key2", "val2")
		lg.Infof(l.name+".Infof xxxx %v", "123")

		lg.Warnln(l.name+".Warn", "key1", "val1", "key2", "val2")
		lg.Warnf(l.name+".Warnf xxxx %v", "123")

		lg.Errorln(l.name+".Error", "key1", "val1", "key2", "val2")
		lg.Errorf(l.name+".Errorf xxxx %v", "123")

		// lg.Fatal(l.name+".Fatal", "key1", "val1", "key2", "val2")
		// lg.Fatalf(l.name+".Fatalf xxxx %v", "123")
	}
}
