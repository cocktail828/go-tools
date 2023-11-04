package log

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

type LogFile struct {
	// Filename is the file to write logs to.  Backup log files will be retained in the same directory.
	// It uses <processname>-lumberjack.log in os.TempDir() if empty.
	Filename string `json:"filename" yaml:"filename"`

	// MaxSize is the maximum size in megabytes of the log file before it gets rotated. It defaults to 100 megabytes.
	MaxSize int `json:"maxsize" yaml:"maxsize"`

	// MaxAge is the maximum number of days to retain old log files based on the timestamp encoded in their filename.
	//  Note that a day is defined as 24 hours and may not exactly correspond to calendar days due to daylight savings,
	// leap seconds, etc. The default is not to remove old log files based on age.
	MaxAge int `json:"maxage" yaml:"maxage"`

	// MaxBackups is the maximum number of old log files to retain.  The default is to retain all old log files (though
	// MaxAge may still cause them to get deleted.)
	MaxBackups int `json:"maxbackups" yaml:"maxbackups"`

	// LocalTime determines if the time used for formatting the timestamps in backup files is the computer's local time.
	// The default is to use UTC time.
	LocalTime bool `json:"localtime" yaml:"localtime"`

	// Compress determines if the rotated log files should be compressed using gzip.
	Compress bool `json:"compress" yaml:"compress"`
}

type Lumberjack struct {
	logger    *lumberjack.Logger
	formatter logrus.Formatter
	levels    []logrus.Level
}

func NewLumberjack(logf LogFile, formatter logrus.Formatter, levels ...logrus.Level) *Lumberjack {
	return &Lumberjack{
		logger: &lumberjack.Logger{
			Filename:   logf.Filename,
			MaxSize:    logf.MaxSize,
			MaxBackups: logf.MaxBackups,
			MaxAge:     logf.MaxAge,
			Compress:   logf.Compress,
			LocalTime:  logf.LocalTime,
		},
		formatter: formatter,
		levels:    levels,
	}
}

func (hook *Lumberjack) Fire(entry *logrus.Entry) error {
	msg, err := hook.formatter.Format(entry)
	if err != nil {
		return err
	}
	_, err = hook.logger.Write([]byte(msg))
	return err
}

func (hook *Lumberjack) Levels() []logrus.Level {
	return hook.levels
}
