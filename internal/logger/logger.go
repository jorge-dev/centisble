package logger

import (
	"context"
	"log/slog"
	"os"
	"runtime"
	"time"

	"github.com/lmittmann/tint"
)

type conditionalSourceHandler struct {
	handler slog.Handler
}

func (h *conditionalSourceHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *conditionalSourceHandler) Handle(ctx context.Context, record slog.Record) error {
	// Check if the log level is Error or Debug
	if record.Level == slog.LevelError || record.Level == slog.LevelDebug {
		// Capture the caller information
		pc, file, line, ok := runtime.Caller(2)
		if ok {
			fn := runtime.FuncForPC(pc)
			record.AddAttrs(
				slog.String("source_file", file),
				slog.Int("source_line", line),
				slog.String("function", fn.Name()),
			)
		}
	}
	return h.handler.Handle(ctx, record)
}

func (h *conditionalSourceHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &conditionalSourceHandler{handler: h.handler.WithAttrs(attrs)}
}

func (h *conditionalSourceHandler) WithGroup(name string) slog.Handler {
	return &conditionalSourceHandler{handler: h.handler.WithGroup(name)}
}

type LogConfig struct {
	Level      slog.Level
	JSONOutput bool
}

func InitLogger(cfg LogConfig) *slog.Logger {
	var handler slog.Handler

	opts := &tint.Options{
		Level:      cfg.Level,
		TimeFormat: time.Kitchen,
	}

	if cfg.JSONOutput {
		// tint does not support JSON output; fallback to default JSON handler
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.Level})
	} else {
		// Use tint's handler for colorized text output
		handler = tint.NewHandler(os.Stdout, opts)
	}

	// Wrap the handler with the conditional source handler
	handler = &conditionalSourceHandler{handler: handler}

	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logger
}
