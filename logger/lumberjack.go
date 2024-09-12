package logger

import (
	"io"
	"log/slog"
	"strings"

	"github.com/cocktail828/go-tools/pkg/lumberjack.v2"
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

	// bufsize define the buffer size, default size 5Mb
	BufSize int `json:"bufsize" toml:"bufsize" yaml:"bufsize" default:"5242880"`

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

func NewLoggerWithLumberjack(cfg Config) *slog.Logger {
	var lvl slog.LevelVar
	lvl.Set(func() slog.Level {
		switch strings.ToLower(cfg.Level) {
		case "debug":
			return slog.LevelDebug
		case "info":
			return slog.LevelInfo
		case "warn":
			return slog.LevelWarn
		case "error":
			return slog.LevelError
		default:
			return slog.LevelError
		}
	}())

	var wr io.Writer
	if cfg.Filename == "/dev/null" {
		wr = io.Discard
	} else {
		wr = &lumberjack.Logger{
			Filename:   cfg.Filename,
			MaxSize:    cfg.MaxSize,
			Async:      cfg.Async,
			BufSize:    cfg.BufSize,
			MaxAge:     cfg.MaxAge,
			MaxBackups: cfg.MaxCount,
			Compress:   cfg.Compress,
		}
	}

	return slog.New(slog.NewJSONHandler(
		wr, &slog.HandlerOptions{
			AddSource: cfg.AddSource,
			Level:     &lvl,
		},
	))
}
