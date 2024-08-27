package logger

import (
	"io"
	"log/slog"
)

func NewNoopLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(
		io.Discard, nil,
	))
}
