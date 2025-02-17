package xlog

import "strings"

// compatible with lumberjack
type Config struct {
	// debug, info, warn, error
	Level string `json:"level" toml:"level" yaml:"level" default:"error"`

	// addsource(ie.. file:line)
	Verbose bool `json:"verbose" toml:"verbose" yaml:"verbose"`

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
