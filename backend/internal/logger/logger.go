package logger

import (
	"log/slog"
	"os"
	"strings"
	"time"
)

// NewLogger создаёт *slog.Logger с JSON-форматом, меткой сервиса и окружения.
func NewLogger(service string, stage string, level string) *slog.Logger {
	var l slog.Level

	switch strings.ToLower(level) {
	case "debug":
		l = slog.LevelDebug
	case "info":
		l = slog.LevelInfo
	case "warn":
		l = slog.LevelWarn
	case "error":
		l = slog.LevelError
	default:
		l = slog.LevelInfo
	}

	logOpts := slog.HandlerOptions{
		Level:       l,
		AddSource:   true,
		ReplaceAttr: replaceAttr,
	}
	logHandler := slog.NewJSONHandler(os.Stdout, &logOpts)

	return slog.New(logHandler).
		With(slog.String("stage", stage)).
		With(slog.String("service", service))
}

// replaceAttr переименовывает стандартные slog-атрибуты в формат,
// совместимый с Cloud Logging / structured logging конвенциями.
func replaceAttr(_ []string, a slog.Attr) slog.Attr {
	switch a.Key {
	case "time":
		return slog.String("timestamp", a.Value.Time().Format(time.RFC3339))
	case "msg":
		return slog.String("rest", a.Value.String())
	case "level":
		return slog.String("severity", a.Value.String())
	default:
		return a
	}
}
