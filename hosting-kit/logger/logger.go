package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

type Level slog.Level

const (
	LevelDebug = Level(slog.LevelDebug)
	LevelInfo  = Level(slog.LevelInfo)
	LevelWarn  = Level(slog.LevelWarn)
	LevelError = Level(slog.LevelError)
)

type TraceIDFn func(ctx context.Context) string

type Logger struct {
	handler   slog.Handler
	traceIDFn TraceIDFn
}

func New(w io.Writer, minLevel Level, serviceName string, traceIDFn TraceIDFn) *Logger {
	if w == nil {
		w = os.Stdout
	}

	replaceAttr := func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.SourceKey {
			if source, ok := a.Value.Any().(*slog.Source); ok {
				v := fmt.Sprintf("%s:%d", filepath.Base(source.File), source.Line)
				return slog.Attr{Key: "file", Value: slog.StringValue(v)}
			}
		}
		return a
	}

	handler := slog.Handler(slog.NewJSONHandler(w, &slog.HandlerOptions{
		AddSource:   true,
		Level:       slog.Level(minLevel),
		ReplaceAttr: replaceAttr,
	}))

	handler = handler.WithAttrs([]slog.Attr{
		slog.String("service", serviceName),
	})

	return &Logger{
		handler:   handler,
		traceIDFn: traceIDFn,
	}
}

func (l *Logger) Handler() slog.Handler {
	return l.handler
}
