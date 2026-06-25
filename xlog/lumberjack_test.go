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

func TestCloseAndRewrite(t *testing.T) {
	l := Logger{
		Filename:   "rewrite.log",
		MaxSize:    1,
		MaxAge:     1,
		MaxBackups: 2,
	}
	defer os.RemoveAll(l.Filename)

	_, err := l.Write([]byte("first write\n"))
	if err != nil {
		t.Fatal(err)
	}

	if err := l.Close(); err != nil {
		t.Fatal(err)
	}

	_, err = l.Write([]byte("second write after close\n"))
	if err != nil {
		t.Fatal(err)
	}

	l.Close()
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
