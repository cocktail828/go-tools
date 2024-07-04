package logger

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

	// AddSource determain whether add file:line to log file.  The default is true
	AddSource bool `json:"addsource" toml:"addsource" yaml:"addsource" default:"true"`

	// Compress determines if the rotated log files should be compressed using gzip.
	Compress bool `json:"compress" toml:"compress" yaml:"compress"`
}

type Logger interface {
	Debugw(msg string, args ...any)
	Debugf(format string, args ...any)
	Infow(msg string, args ...any)
	Infof(format string, args ...any)
	Warnw(msg string, args ...any)
	Warnf(format string, args ...any)
	Errorw(msg string, args ...any)
	Errorf(format string, args ...any)
	With(args ...any) Logger
	WithGroup(name string) Logger
}
