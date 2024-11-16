package cron

import (
	"fmt"
	"log/slog"
)

type cronLogger struct{}

func (cl cronLogger) Info(msg string, keysAndValues ...any) {
	slog.Info(fmt.Sprintf("Cron: %s", msg), keysAndValues...)
}

func (cl cronLogger) Error(err error, msg string, keysAndValues ...any) {
	slog.Error(fmt.Sprintf("Cron: %s: %v", msg, err), keysAndValues...)
}
