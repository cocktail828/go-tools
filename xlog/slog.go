package xlog

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"runtime"
	"time"
)

type WrapperLogger struct {
	logger *slog.Logger
	slog.HandlerOptions
}

func New(h slog.Handler, opts slog.HandlerOptions) Logger {
	return &WrapperLogger{slog.New(h), opts}
}

func NewJSONHandler(w io.Writer, opts HandlerOptions) Logger {
	sopts := slog.HandlerOptions{
		AddSource: opts.AddSource,
		Level:     opts.Level,
	}
	return &WrapperLogger{slog.New(slog.NewJSONHandler(w, &sopts)), sopts}
}

func NewTextHandler(w io.Writer, opts HandlerOptions) Logger {
	sopts := slog.HandlerOptions{
		AddSource: opts.AddSource,
		Level:     opts.Level,
	}
	return &WrapperLogger{slog.New(slog.NewTextHandler(w, &sopts)), sopts}
}

func (wl *WrapperLogger) With(args ...any) Logger {
	return &WrapperLogger{
		logger:         wl.logger.With(args...),
		HandlerOptions: wl.HandlerOptions,
	}
}

func (wl *WrapperLogger) WithGroup(name string) Logger {
	return &WrapperLogger{
		logger:         wl.logger.WithGroup(name),
		HandlerOptions: wl.HandlerOptions,
	}
}

func (wl *WrapperLogger) log(level slog.Level, msg string, args ...any) {
	if !wl.logger.Enabled(context.Background(), level) {
		return
	}
	var pc uintptr
	if wl.AddSource {
		var pcs [1]uintptr
		// skip [runtime.Callers, this function, this function's caller]
		runtime.Callers(3, pcs[:])
		pc = pcs[0]
	}
	r := slog.NewRecord(time.Now(), level, msg, pc)
	r.Add(args...)
	_ = wl.logger.Handler().Handle(context.Background(), r)
}

func (wl *WrapperLogger) logf(level slog.Level, format string, args ...any) {
	if !wl.logger.Enabled(context.Background(), level) {
		return
	}
	var pc uintptr
	if wl.AddSource {
		var pcs [1]uintptr
		// skip [runtime.Callers, this function, this function's caller]
		runtime.Callers(3, pcs[:])
		pc = pcs[0]
	}
	r := slog.NewRecord(time.Now(), level, fmt.Sprintf(format, args...), pc)
	_ = wl.logger.Handler().Handle(context.Background(), r)
}

func (wl *WrapperLogger) Debugln(msg string, args ...any) {
	wl.log(slog.LevelDebug, msg, args...)
}

func (wl *WrapperLogger) Debugf(format string, args ...any) {
	wl.logf(slog.LevelDebug, format, args...)
}

func (wl *WrapperLogger) Infoln(msg string, args ...any) {
	wl.log(slog.LevelInfo, msg, args...)
}

func (wl *WrapperLogger) Infof(format string, args ...any) {
	wl.logf(slog.LevelInfo, format, args...)
}

func (wl *WrapperLogger) Warnln(msg string, args ...any) {
	wl.log(slog.LevelWarn, msg, args...)
}

func (wl *WrapperLogger) Warnf(format string, args ...any) {
	wl.logf(slog.LevelWarn, format, args...)
}

func (wl *WrapperLogger) Errorln(msg string, args ...any) {
	wl.log(slog.LevelError, msg, args...)
}

func (wl *WrapperLogger) Errorf(format string, args ...any) {
	wl.logf(slog.LevelError, format, args...)
}
