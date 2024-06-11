package logger

import (
	"fmt"

	"golang.org/x/exp/slog"
)

// slogWrapper implements Logger interface
type slogWrapper struct {
	l *slog.Logger
}

// NewLoggerWithSlog creates a new logger which wraps
// the given logrus.Logger
func NewLoggerWithSlog(logger *slog.Logger) Logger {
	return slogWrapper{
		l: logger,
	}
}

func (c slogWrapper) Debugw(msg string, args ...any)    { c.l.Debug(msg, args...) }
func (c slogWrapper) Debugf(format string, args ...any) { c.l.Debug(fmt.Sprintf(format, args...)) }

func (c slogWrapper) Infow(msg string, args ...any)    { c.l.Info(msg, args...) }
func (c slogWrapper) Infof(format string, args ...any) { c.l.Info(fmt.Sprintf(format, args...)) }

func (c slogWrapper) Warnw(msg string, args ...any)    { c.l.Warn(msg, args...) }
func (c slogWrapper) Warnf(format string, args ...any) { c.l.Warn(fmt.Sprintf(format, args...)) }

func (c slogWrapper) Errorw(msg string, args ...any)    { c.l.Error(msg, args...) }
func (c slogWrapper) Errorf(format string, args ...any) { c.l.Error(fmt.Sprintf(format, args...)) }

func (c slogWrapper) With(args ...any) Logger {
	return NewLoggerWithSlog(c.l.With(args...))
}

func (c slogWrapper) WithGroup(name string) Logger {
	return NewLoggerWithSlog(c.l.WithGroup(name))
}
