package pkg

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"sync"
)

const (
	// Time
	timeFormat = "[15:04:05.000]"
	// Utils
	reset = "\033[0m"
	// Colors
	black        = 30
	red          = 31
	green        = 32
	yellow       = 33
	blue         = 34
	magenta      = 35
	cyan         = 36
	lightGray    = 37
	darkGray     = 90
	lightRed     = 91
	lightGreen   = 92
	lightYellow  = 93
	lightBlue    = 94
	lightMagenta = 95
	lightCyan    = 96
	white        = 97
)

func colorize(colorCode int, v string) string {
	return fmt.Sprintf("\033[%sm%s%s", strconv.Itoa(colorCode), v, reset)
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

func NewHandler(opts *slog.HandlerOptions) *LogHandler {
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

func SetupLogger(cfg LogConfig) {

	var logger slog.Handler
	switch cfg.Format {
	case "text":
		logger = NewHandler(nil)
	case "json":
		logger = NewHandler(&slog.HandlerOptions{
			Level: slog.LevelDebug,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == "nothing" {
					return slog.Attr{}
				}
				return a
			},
		})
		logger.WithGroup("data")
	default:
		slog.Error("Unknown log format", slog.String("format", cfg.Format))
		os.Exit(-1)
	}
	slog.SetDefault(slog.New(logger))
}

type LogHandler struct {
	handler slog.Handler
	buffer  *bytes.Buffer
	mutex   *sync.Mutex
}

func (lh *LogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return lh.handler.Enabled(ctx, level)
}

func (lh *LogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &LogHandler{handler: lh.handler.WithAttrs(attrs), buffer: lh.buffer, mutex: lh.mutex}
}

func (lh *LogHandler) WithGroup(name string) slog.Handler {
	return &LogHandler{handler: lh.handler.WithGroup(name), buffer: lh.buffer, mutex: lh.mutex}
}

func (lh *LogHandler) Handle(ctx context.Context, r slog.Record) error {
	level := r.Level.String() + ":"

	switch r.Level {
	case slog.LevelDebug:
		level = colorize(darkGray, level)
	case slog.LevelInfo:
		level = colorize(cyan, level)
	case slog.LevelWarn:
		level = colorize(lightYellow, level)
	case slog.LevelError:
		level = colorize(lightRed, level)
	}

	attrs, err := lh.computeAttrs(ctx, r)
	if err != nil {
		return err
	}

	bytes, err := json.MarshalIndent(attrs, "", "  ")
	if err != nil {
		return fmt.Errorf("error when marshaling attrs: %w", err)
	}

	fmt.Println(
		colorize(lightGray, r.Time.Format(timeFormat)),
		level,
		colorize(white, strings.TrimSuffix(r.Message, "\n")),
	)

	if len(bytes) > 2 {
		fmt.Printf("%s\n",
			colorize(darkGray, string(bytes)),
		)
	}

	return nil
}

func (lh *LogHandler) computeAttrs(ctx context.Context, r slog.Record) (map[string]any, error) {
	lh.mutex.Lock()
	defer func() {
		lh.buffer.Reset()
		lh.mutex.Unlock()
	}()
	if err := lh.handler.Handle(ctx, r); err != nil {
		return nil, fmt.Errorf("error when calling inner handler's Handle: %w", err)
	}

	var attrs map[string]any
	err := json.Unmarshal(lh.buffer.Bytes(), &attrs)
	if err != nil {
		return nil, fmt.Errorf("error when unmarshaling inner handler's Handle result: %w", err)
	}
	return attrs, nil
}

type CronLogger struct{}

func (cl CronLogger) Info(msg string, keysAndValues ...any) {
	slog.Info(fmt.Sprintf("Cron: %s", msg), keysAndValues...)
}

func (cl CronLogger) Error(err error, msg string, keysAndValues ...any) {
	slog.Error(fmt.Sprintf("Cron: %s: %v", msg, err), keysAndValues...)
}
