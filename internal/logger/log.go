package logger

import (
	"log/slog"
	"time"
)

func Info(opt Operation, start time.Time, msg string, v ...any) {
	slog.Info(msg, append([]any{"operation", opt.ToString(), "duration"}, v...)...)
}
