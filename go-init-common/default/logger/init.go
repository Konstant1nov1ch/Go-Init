package logger

import (
	"log/slog"
	"os"
)

// New создает новый экземпляр логгера.
func New(cfg *Config, serviceName, env string) *Logger {
	var level slog.Level
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		level = slog.LevelInfo
	}

	var handler slog.Handler
	switch cfg.Format {
	case "native":
		handler = textLogger(level)
	case "json":
		handler = jsonLogger(level)
	default:
		handler = jsonLogger(level)
	}

	handler = handler.WithAttrs([]slog.Attr{
		String("serviceName", serviceName),
		String("env", env),
	})

	logger := slog.New(&ContextHandler{handler})
	slog.SetDefault(logger)

	logger.Info("Logger initialized", String("Level", cfg.Level), String("Format", cfg.Format))
	return &Logger{Log: logger}
}

func jsonLogger(loglvl slog.Level) slog.Handler {
	return slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: loglvl})
}

func textLogger(loglvl slog.Level) slog.Handler {
	return slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: loglvl})
}
