package xlog

import (
	"os"
	"testing"
)

func TestNoCache(t *testing.T) {
	cfg := Config{
		Filename:   "no-cache.log",
		MaxSize:    100,
		MaxAge:     1,
		MaxBackups: 3,
	}
	defer os.RemoveAll(cfg.Filename)

	l := cfg.Writer()
	l.Write([]byte("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx\n"))
}

func TestWithCache(t *testing.T) {
	cfg := Config{
		BufSize:    1024 * 1024 * 10,
		Filename:   "cache.log",
		MaxSize:    100,
		MaxAge:     1,
		MaxBackups: 3,
	}
	defer os.RemoveAll(cfg.Filename)

	l := cfg.Writer()
	l.Write([]byte("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx\n"))
}

func BenchmarkLumberjack(b *testing.B) {
	cfg := Config{
		BufSize:    1024 * 1024 * 10,
		Filename:   "/log/server/xxx.log",
		MaxSize:    100,
		MaxAge:     1,
		MaxBackups: 3,
	}
	defer os.RemoveAll(cfg.Filename)

	b.Run("no-cache", func(b *testing.B) {
		w := cfg.Writer()
		b.ResetTimer()
		b.RunParallel(func(p *testing.PB) {
			for p.Next() {
				w.Write([]byte("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx\n"))
			}
		})
	})
}
