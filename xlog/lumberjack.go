package xlog

import (
	"bufio"
	"io"
	"strings"
	"sync"

	"gopkg.in/natefinch/lumberjack.v2"
)

// Config is the configuration for lumberjack logger.
type Config struct {
	// debug, info, warn, error
	Level string `hcl:"level" json:"level" toml:"level" yaml:"level" default:"error"`

	// addsource(ie.. file:line)
	Verbose bool `hcl:"verbose" json:"verbose" toml:"verbose" yaml:"verbose"`

	// BufSize is the buffer size in bytes
	BufSize int `hcl:"bufsize" json:"bufsize" toml:"bufsize" yaml:"bufsize"`

	// Filename is the file to write logs to.  Backup log files will be retained in the same directory.
	// It uses <processname>-lumberjack.log in os.TempDir() if empty.
	Filename string `hcl:"filename" json:"filename" toml:"filename" yaml:"filename" validate:"required"`

	// MaxSize is the maximum size in megabytes of the log file before it gets rotated. It defaults to 100 megabytes.
	MaxSize int `hcl:"maxsize" json:"maxsize" toml:"maxsize" yaml:"maxsize" default:"100"`

	// MaxAge is the maximum number of days to retain old log files based on the timestamp encoded in their filename.
	//  Note that a day is defined as 24 hours and may not exactly correspond to calendar days due to daylight savings,
	// leap seconds, etc. The default is not to remove old log files based on age.
	MaxAge int `hcl:"maxage" json:"maxage" toml:"maxage" yaml:"maxage" default:"7"`

	// MaxBackups is the maximum number of old log files to retain.  The default is to retain all old log files (though
	// MaxAge may still cause them to get deleted.)
	MaxBackups int `hcl:"maxbackups" json:"maxbackups" toml:"maxbackups" yaml:"maxbackups" default:"5"`
}

func (cfg Config) GetLevel() Level {
	lvl := strings.ToLower(cfg.Level)
	switch {
	case strings.Contains(lvl, "debug"):
		return LevelDebug
	case strings.Contains(lvl, "info"):
		return LevelInfo
	case strings.Contains(lvl, "warn"):
		return LevelWarn
	case strings.Contains(lvl, "error"):
		fallthrough
	default:
		return LevelError
	}
}

type bufferWriter struct {
	mu sync.Mutex
	wr io.Writer
}

func (w *bufferWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.wr.Write(p)
}

func (cfg *Config) Writer() io.Writer {
	if cfg.Filename == "/dev/null" || cfg.Filename == "" {
		return io.Discard
	}

	var w io.Writer = &lumberjack.Logger{
		Filename:   cfg.Filename,
		MaxSize:    cfg.MaxSize,
		MaxAge:     cfg.MaxAge,
		MaxBackups: cfg.MaxBackups,
	}

	if cfg.BufSize > 0 {
		w = &bufferWriter{
			wr: bufio.NewWriterSize(w, cfg.BufSize),
		}
	}

	return w
}
