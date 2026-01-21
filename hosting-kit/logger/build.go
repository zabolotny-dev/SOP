package logger

import (
	"context"
	"runtime/debug"
)

func (l *Logger) BuildInfo(ctx context.Context) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		l.Warn(ctx, "build info not available")
		return
	}

	var args []any

	for _, s := range info.Settings {
		args = append(args, s.Key, s.Value)
	}

	args = append(args, "goversion", info.GoVersion)
	args = append(args, "modversion", info.Main.Version)

	l.Info(ctx, "build info", args...)
}
