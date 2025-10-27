package colorful

import (
	"os"
	"testing"
	"time"

	"github.com/cocktail828/go-tools/xlog"
	"golang.org/x/time/rate"
)

func TestColorful(t *testing.T) {
	logs := []func() *Logger{
		func() *Logger {
			l := Default()
			l.SetPrefix("[default] ")
			l.SetFlags(l.Flags() | Llongfile)
			return l
		},
		func() *Logger {
			l := NewColorful(os.Stderr, "", LstdFlags|Llongfile)
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

	{
		l := Default().Limited(rate.NewLimiter(rate.Every(time.Second), 1))
		for range 100 {
			l.Errorln("Error", "key1", "val1", "key2", "val2")
			time.Sleep(time.Millisecond * 10)
		}
	}
}
