package xlog

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime"
	"time"
)

type Logger struct {
	logger *slog.Logger
	slog.HandlerOptions
}

func New(h slog.Handler, opts slog.HandlerOptions) *Logger {
	return &Logger{slog.New(h), opts}
}

func NewJSONHandler(w io.Writer, opts HandlerOptions) *Logger {
	sopts := slog.HandlerOptions{
		AddSource: opts.AddSource,
		Level:     opts.Level,
	}
	return &Logger{slog.New(slog.NewJSONHandler(w, &sopts)), sopts}
}

func NewTextHandler(w io.Writer, opts HandlerOptions) *Logger {
	sopts := slog.HandlerOptions{
		AddSource: opts.AddSource,
		Level:     opts.Level,
	}
	return &Logger{slog.New(slog.NewTextHandler(w, &sopts)), sopts}
}

func (l *Logger) With(args ...any) *Logger {
	return &Logger{
		logger:         l.logger.With(args...),
		HandlerOptions: l.HandlerOptions,
	}
}

func (l *Logger) WithGroup(name string) *Logger {
	return &Logger{
		logger:         l.logger.WithGroup(name),
		HandlerOptions: l.HandlerOptions,
	}
}

func (l *Logger) log(level slog.Level, msg string, args ...any) {
	if !l.logger.Enabled(context.Background(), level) {
		return
	}
	var pc uintptr
	if l.AddSource {
		var pcs [1]uintptr
		// skip [runtime.Callers, this function, this function's caller]
		runtime.Callers(3, pcs[:])
		pc = pcs[0]
	}
	r := slog.NewRecord(time.Now(), level, msg, pc)
	r.Add(args...)
	_ = l.logger.Handler().Handle(context.Background(), r)
}

// always dump log, ignore log level
func (l *Logger) alwayslog(level slog.Level, msg string, args ...any) {
	var pc uintptr
	if l.AddSource {
		var pcs [1]uintptr
		// skip [runtime.Callers, this function, this function's caller]
		runtime.Callers(3, pcs[:])
		pc = pcs[0]
	}
	r := slog.NewRecord(time.Now(), level, msg, pc)
	r.Add(args...)
	_ = l.logger.Handler().Handle(context.Background(), r)
}

func (l *Logger) Println(msg string, args ...any) {
	l.alwayslog(slog.LevelInfo, msg, args...)
}

func (l *Logger) Printf(format string, args ...any) {
	l.alwayslog(slog.LevelInfo, fmt.Sprintf(format, args...))
}

func (l *Logger) Debugln(msg string, args ...any) {
	l.log(slog.LevelDebug, msg, args...)
}

func (l *Logger) Debugf(format string, args ...any) {
	l.log(slog.LevelDebug, fmt.Sprintf(format, args...))
}

func (l *Logger) Infoln(msg string, args ...any) {
	l.log(slog.LevelInfo, msg, args...)
}

func (l *Logger) Infof(format string, args ...any) {
	l.log(slog.LevelInfo, fmt.Sprintf(format, args...))
}

func (l *Logger) Warnln(msg string, args ...any) {
	l.log(slog.LevelWarn, msg, args...)
}

func (l *Logger) Warnf(format string, args ...any) {
	l.log(slog.LevelWarn, fmt.Sprintf(format, args...))
}

func (l *Logger) Errorln(msg string, args ...any) {
	l.log(slog.LevelError, msg, args...)
}

func (l *Logger) Errorf(format string, args ...any) {
	l.log(slog.LevelError, fmt.Sprintf(format, args...))
}

func (l *Logger) Fatalln(msg string, args ...any) {
	l.alwayslog(slog.LevelError, msg, args...)
	os.Exit(1)
}

func (l *Logger) Fatalf(format string, args ...any) {
	l.alwayslog(slog.LevelError, fmt.Sprintf(format, args...))
	os.Exit(1)
}
