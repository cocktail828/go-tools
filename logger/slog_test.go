package logger_test

import (
	"testing"

	"github.com/cocktail828/go-tools/configor"
	"github.com/cocktail828/go-tools/logger"
	"github.com/cocktail828/go-tools/z"
)

func TestSlog(t *testing.T) {
	l := logger.NewLoggerWithLumberjack(logger.Config{
		Level:     "error",
		Filename:  "/log/server/error.log",
		MaxSize:   100,
		MaxCount:  1,
		MaxAge:    1,
		AddSource: true,
		Compress:  false,
	})
	l = l.With("a1", "b1").WithGroup("xxx")
	l.Infow("slog.finished", "key", "value")
	l.Errorw("slog.finishedxxx", "key", "value")
	l.Errorf("slog.finishedxxx %v", "key")
}

func BenchmarkLog(b *testing.B) {
	cfg := logger.Config{}
	z.Must(configor.Load(&cfg, []byte(`
level = "debug"
filename = "/log/server/xxx.log"
async = true
`)))

	b.Run("no-cache", func(b *testing.B) {
		cfg.Async = false
		l := logger.NewLoggerWithLumberjack(cfg)
		b.ResetTimer()
		b.RunParallel(func(p *testing.PB) {
			for p.Next() {
				l.Errorf("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx")
			}
		})
	})

	b.Run("cache", func(b *testing.B) {
		cfg.Async = true
		l := logger.NewLoggerWithLumberjack(cfg)
		b.ResetTimer()
		b.RunParallel(func(p *testing.PB) {
			for p.Next() {
				l.Errorf("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx")
			}
		})
	})
}
