package commands

import (
	"context"
	"feedexampleredis/internal/storage"
	"fmt"
	"log/slog"
	"os"
)

func InsertFeedFile(ctx context.Context, path string, redisClient *storage.Redis) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	// Insert the feed
	count, err := redisClient.StreamingFeedInsert(ctx, f)
	if err != nil {
		return fmt.Errorf("failed to insert feed: %w", err)
	}

	slog.Info(
		"feed inserted",
		slog.Int64("count", count),
	)

	return nil
}
