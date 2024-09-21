package sdk

import (
	"log/slog"
	"os"
)

func SetupLogger(cfg LogConfig) {
	opts := &slog.HandlerOptions{
		AddSource: cfg.AddSource,
		Level:     cfg.Level,
	}

	var logger slog.Handler
	switch cfg.Format {
	case "json":
		logger = slog.NewJSONHandler(os.Stdout, opts)
	case "text":
		logger = slog.NewTextHandler(os.Stdout, opts)
	default:
		slog.Error("Unknown log format", slog.String("format", cfg.Format))
		os.Exit(-1)
	}
	slog.SetDefault(slog.New(logger))
}
