package logger

import (
	"context"
	"log/slog"
)

// Logger управляет логированием в системе.
type Logger struct {
	Log *slog.Logger
}

func (l *Logger) Info(message string, args ...any)  { l.Log.Info(message, args...) }
func (l *Logger) Debug(message string, args ...any) { l.Log.Debug(message, args...) }
func (l *Logger) Error(message string, args ...any) { l.Log.Error(message, args...) }
func (l *Logger) Warn(message string, args ...any)  { l.Log.Warn(message, args...) }
func (l *Logger) InfoContext(ctx context.Context, message string, args ...any) {
	l.Log.InfoContext(ctx, message, args...)
}

func (l *Logger) DebugContext(ctx context.Context, message string, args ...any) {
	l.Log.DebugContext(ctx, message, args...)
}

func (l *Logger) WarnContext(ctx context.Context, message string, args ...any) {
	l.Log.WarnContext(ctx, message, args...)
}

func (l *Logger) ErrorContext(ctx context.Context, message string, args ...any) {
	l.Log.ErrorContext(ctx, message, args...)
}
