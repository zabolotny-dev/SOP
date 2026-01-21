package logger

import (
	"context"
	"log/slog"
	"runtime"
	"time"
)

func (l *Logger) Debug(ctx context.Context, msg string, args ...any) {
	l.log(ctx, LevelDebug, msg, args...)
}

func (l *Logger) Info(ctx context.Context, msg string, args ...any) {
	l.log(ctx, LevelInfo, msg, args...)
}

func (l *Logger) Warn(ctx context.Context, msg string, args ...any) {
	l.log(ctx, LevelWarn, msg, args...)
}

func (l *Logger) Error(ctx context.Context, msg string, args ...any) {
	l.log(ctx, LevelError, msg, args...)
}

func (l *Logger) log(ctx context.Context, level Level, msg string, args ...any) {
	slogLevel := slog.Level(level)

	if !l.handler.Enabled(ctx, slogLevel) {
		return
	}

	var pcs [1]uintptr
	runtime.Callers(3, pcs[:])

	r := slog.NewRecord(time.Now(), slogLevel, msg, pcs[0])

	if l.traceIDFn != nil {
		if traceID := l.traceIDFn(ctx); traceID != "" && traceID != "00000000000000000000000000000000" {
			args = append(args, "trace_id", traceID)
		}
	}

	r.Add(args...)

	l.handler.Handle(ctx, r)
}
