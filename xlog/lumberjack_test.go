package xlog

import (
	"os"
	"testing"
)

func TestNoCache(t *testing.T) {
	l := Logger{
		Filename:   "no-cache.log",
		MaxSize:    100,
		MaxAge:     1,
		MaxBackups: 3,
	}
	defer os.RemoveAll(l.Filename)

	l.Write([]byte("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx\n"))
}

func TestWithCache(t *testing.T) {
	l := Logger{
		BufSize:    10,
		Filename:   "cache.log",
		MaxSize:    100,
		MaxAge:     1,
		MaxBackups: 2,
	}

	// defer os.RemoveAll(l.Filename)
	defer l.Close()

	for range 100_0000 {
		l.Write([]byte("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx\n"))
	}
}

func BenchmarkLumberjack(b *testing.B) {
	l := Logger{
		BufSize:    1024 * 1024 * 10,
		Filename:   "/log/server/xxx.log",
		MaxSize:    100,
		MaxAge:     1,
		MaxBackups: 3,
	}
	defer os.RemoveAll(l.Filename)

	b.Run("no-cache", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(p *testing.PB) {
			for p.Next() {
				l.Write([]byte("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx\n"))
			}
		})
	})
}
