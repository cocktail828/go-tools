package lumberjack

import (
	"bufio"
	"io"
	"sync"

	"github.com/cocktail828/go-tools/z/environ"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Config struct {
	// Filename is the file to write logs to.  Backup log files will be retained in the same directory.
	// It uses <processname>-lumberjack.log in os.TempDir() if empty.
	Filename string `json:"filename" toml:"filename" yaml:"filename" validate:"required"`

	// MaxSize is the maximum size in megabytes of the log file before it gets rotated. It defaults to 100 megabytes.
	MaxSize int `json:"maxsize" toml:"maxsize" yaml:"maxsize" default:"100"`

	// Async will cache log and flush on need(30s timeout or buffer is full)
	// `XLOG_BUF_SIZE` define the buffer size
	Async bool `json:"async" toml:"async" yaml:"async"`

	// MaxAge is the maximum number of days to retain old log files based on the timestamp encoded in their filename.
	//  Note that a day is defined as 24 hours and may not exactly correspond to calendar days due to daylight savings,
	// leap seconds, etc. The default is not to remove old log files based on age.
	MaxAge int `json:"maxage" toml:"maxage" yaml:"maxage" default:"7"`

	// MaxCount is the maximum number of old log files to retain.  The default is to retain all old log files (though
	// MaxAge may still cause them to get deleted.)
	MaxCount int `json:"maxcount" toml:"maxcount" yaml:"maxcount" default:"5"`

	// Compress determines if the rotated log files should be compressed using gzip.
	Compress bool `json:"compress" toml:"compress" yaml:"compress"`
}

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

func NewWriter(cfg Config) io.Writer {
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
		if val := int(environ.Int64("XLOG_BUF_SIZE")); val > MinBufSize {
			bufsize = val
		}
		w = &bufferWriter{
			wr: bufio.NewWriterSize(w, bufsize),
		}
	}

	return w
}
