package lumberjack

import (
	"bufio"
	"io"
	"log/slog"
	"sync"

	"github.com/cocktail828/go-tools/xlog"
	"github.com/cocktail828/go-tools/z/environs"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Config struct {
	// log level, debug, info, warn, error
	Level string `json:"level" toml:"level" yaml:"level" validate:"required"`

	// Filename is the file to write logs to.  Backup log files will be retained in the same directory.
	// It uses <processname>-lumberjack.log in os.TempDir() if empty.
	Filename string `json:"filename" toml:"filename" yaml:"filename" validate:"required"`

	// MaxSize is the maximum size in megabytes of the log file before it gets rotated. It defaults to 100 megabytes.
	MaxSize int `json:"maxsize" toml:"maxsize" yaml:"maxsize" default:"100"`

	// Async will cache log and flush on need(30s timeout or buffer is full)
	Async bool `json:"async" toml:"async" yaml:"async"`

	// MaxAge is the maximum number of days to retain old log files based on the timestamp encoded in their filename.
	//  Note that a day is defined as 24 hours and may not exactly correspond to calendar days due to daylight savings,
	// leap seconds, etc. The default is not to remove old log files based on age.
	MaxAge int `json:"maxage" toml:"maxage" yaml:"maxage" default:"7"`

	// MaxCount is the maximum number of old log files to retain.  The default is to retain all old log files (though
	// MaxAge may still cause them to get deleted.)
	MaxCount int `json:"maxcount" toml:"maxcount" yaml:"maxcount" default:"5"`

	// AddSource determain whether add file:line to log file.
	AddSource bool `json:"addsource" toml:"addsource" yaml:"addsource"`

	// Compress determines if the rotated log files should be compressed using gzip.
	Compress bool `json:"compress" toml:"compress" yaml:"compress"`
}

var (
	MinBufSize = 10 * 1024 * 1024 // 10MB
)

type BufferWriter struct {
	mu sync.RWMutex
	wr io.Writer
}

func (w *BufferWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.wr.Write(p)
}

func NewLumberjack(cfg Config) *xlog.Logger {
	var w io.Writer
	if cfg.Filename == "/dev/null" {
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
		if val := int(environs.Int64("XLOG_BUF_SIZE")); val > MinBufSize {
			bufsize = val
		}
		w = &BufferWriter{
			wr: bufio.NewWriterSize(w, bufsize),
		}
	}

	level := slog.LevelError
	level.UnmarshalText([]byte(cfg.Level))

	sopts := slog.HandlerOptions{
		AddSource: cfg.AddSource,
		Level:     level,
	}
	return xlog.New(slog.NewJSONHandler(w, &sopts), sopts)
}
