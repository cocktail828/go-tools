package lumberjack

import (
	"bufio"
	"io"
	"sync"

	"github.com/cocktail828/go-tools/xlog"
	"github.com/cocktail828/go-tools/z/environ"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	MinBufSize = 10 * 1024 * 1024 // 10MB
)

type bufferWriter struct {
	mu sync.Mutex
	wr io.Writer
}

func (w *bufferWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.wr.Write(p)
}

func NewWriter(cfg xlog.Config) io.Writer {
	var w io.Writer
	if cfg.Filename == "/dev/null" || cfg.Filename == "" {
		w = io.Discard
	} else {
		w = &lumberjack.Logger{
			Filename:   cfg.Filename,
			MaxSize:    cfg.MaxSize,
			MaxAge:     cfg.MaxAge,
			MaxBackups: cfg.MaxCount,
			Compress:   cfg.Compress,
		}
	}

	if cfg.Async {
		bufsize := MinBufSize
		if val := int(environ.Int("XLOG_BUF_SIZE")); val > MinBufSize {
			bufsize = val
		}
		w = &bufferWriter{
			wr: bufio.NewWriterSize(w, bufsize),
		}
	}

	return w
}
