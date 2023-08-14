package commands

import (
	"context"
	"feedexampleredis/internal/app"
	"feedexampleredis/internal/storage"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"
)

func MergeRealtimeFile(ctx context.Context, cfg app.Config, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	// Setup the redis client
	ttl := time.Duration(cfg.TTL) * time.Hour
	redisClient := storage.NewRedis(cfg.RedisAddr, cfg.RedisPass, cfg.RedisDB, ttl, cfg.ConcurrentNum, cfg.ChunkSize)
	redisClient.Connect()
	slog.Info(
		"redis client created",
		slog.String("redis_addr", cfg.RedisAddr),
		slog.String("redis_db", strconv.Itoa(cfg.RedisDB)),
	)

	// Insert the feed
	count, err := redisClient.StreamingMergeInsert(ctx, f)
	if err != nil {
		return fmt.Errorf("failed to insert feed: %w", err)
	}

	slog.Info(
		"feed inserted",
		slog.Int64("count", count),
	)

	return nil
}
