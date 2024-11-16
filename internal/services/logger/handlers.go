package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"
)

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
