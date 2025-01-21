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

// always dump log, ignore log level
func (wl *WrapperLogger) alwayslog(level slog.Level, msg string, args ...any) {
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

func (wl *WrapperLogger) Println(msg string, args ...any) {
	wl.alwayslog(slog.LevelInfo, msg, args...)
}

func (wl *WrapperLogger) Printf(format string, args ...any) {
	wl.alwayslog(slog.LevelInfo, fmt.Sprintf(format, args...))
}

func (wl *WrapperLogger) Debugln(msg string, args ...any) {
	wl.log(slog.LevelDebug, msg, args...)
}

func (wl *WrapperLogger) Debugf(format string, args ...any) {
	wl.log(slog.LevelDebug, fmt.Sprintf(format, args...))
}

func (wl *WrapperLogger) Infoln(msg string, args ...any) {
	wl.log(slog.LevelInfo, msg, args...)
}

func (wl *WrapperLogger) Infof(format string, args ...any) {
	wl.log(slog.LevelInfo, fmt.Sprintf(format, args...))
}

func (wl *WrapperLogger) Warnln(msg string, args ...any) {
	wl.log(slog.LevelWarn, msg, args...)
}

func (wl *WrapperLogger) Warnf(format string, args ...any) {
	wl.log(slog.LevelWarn, fmt.Sprintf(format, args...))
}

func (wl *WrapperLogger) Errorln(msg string, args ...any) {
	wl.log(slog.LevelError, msg, args...)
}

func (wl *WrapperLogger) Errorf(format string, args ...any) {
	wl.log(slog.LevelError, fmt.Sprintf(format, args...))
}

func (wl *WrapperLogger) Fatalln(msg string, args ...any) {
	wl.alwayslog(slog.LevelError, msg, args...)
	os.Exit(1)
}

func (wl *WrapperLogger) Fatalf(format string, args ...any) {
	wl.alwayslog(slog.LevelError, fmt.Sprintf(format, args...))
	os.Exit(1)
}
