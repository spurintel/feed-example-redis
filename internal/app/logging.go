package app

import (
	"os"

	"log/slog"
)

func InitLogging(version, commit, date string) {
	// Configure slog
	level := slog.LevelInfo
	envLevel := os.Getenv("SPUR_REDIS_LOG_LEVEL")
	switch envLevel {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	textHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
	}).WithAttrs([]slog.Attr{
		slog.String("app", "feedexampleredis"),
		slog.String("version", version),
		slog.String("commit", commit),
		slog.String("build_date", date),
	})

	logger := slog.New(textHandler)
	slog.SetDefault(logger)
}
