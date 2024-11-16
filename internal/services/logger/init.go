package logger

import (
	"bytes"
	"log/slog"
	"os"
	"sync"
)

var (
	Logger *slog.Logger
)

func Init(level slog.Level, format string, addSource bool) *slog.Logger {
	sync.OnceFunc(func() {
		var logger slog.Handler
		switch format {
		case "text":
			logger = newHandler(&slog.HandlerOptions{
				Level:     level,
				AddSource: addSource,
			})
		case "json":
			logger = newHandler(&slog.HandlerOptions{
				Level:     level,
				AddSource: addSource,
				ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
					if a.Key == "nothing" {
						return slog.Attr{}
					}
					return a
				},
			})
			logger.WithGroup("data")
		default:
			slog.Error("Unknown log format", slog.String("format", format))
			os.Exit(-1)
		}
		Logger = slog.New(logger)
		slog.SetDefault(Logger)
	})()

	return Logger
}

func newHandler(opts *slog.HandlerOptions) *LogHandler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}
	buffer := &bytes.Buffer{}
	return &LogHandler{
		buffer: buffer,
		handler: slog.NewJSONHandler(buffer, &slog.HandlerOptions{
			Level:       opts.Level,
			AddSource:   opts.AddSource,
			ReplaceAttr: suppressDefaults(opts.ReplaceAttr),
		}),
		mutex: &sync.Mutex{},
	}
}

func suppressDefaults(next func([]string, slog.Attr) slog.Attr) func([]string, slog.Attr) slog.Attr {
	return func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.TimeKey ||
			a.Key == slog.LevelKey ||
			a.Key == slog.MessageKey {
			return slog.Attr{}
		}
		if next == nil {
			return a
		}
		return next(groups, a)
	}
}
