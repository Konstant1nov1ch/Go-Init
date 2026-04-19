package logger

import (
	"log/slog"
	"time"
)

const errMsg = "error-message"

// Атрибуты для логов.
func Any(key string, val interface{}) slog.Attr        { return slog.Any(key, val) }
func String(key, val string) slog.Attr                 { return slog.String(key, val) }
func Bool(key string, val bool) slog.Attr              { return slog.Bool(key, val) }
func Int(key string, val int) slog.Attr                { return slog.Int(key, val) }
func Int64(key string, val int64) slog.Attr            { return slog.Int64(key, val) }
func Uint64(key string, val uint64) slog.Attr          { return slog.Uint64(key, val) }
func Float64(key string, val float64) slog.Attr        { return slog.Float64(key, val) }
func Time(key string, val time.Time) slog.Attr         { return slog.Time(key, val) }
func Duration(key string, val time.Duration) slog.Attr { return slog.Duration(key, val) }
func Error(err error) slog.Attr                        { return slog.String("error-message", err.Error()) }
