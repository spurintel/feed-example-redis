package commands

import (
	"context"
	"feedexampleredis/internal/app"
	"feedexampleredis/internal/spur"
	"feedexampleredis/internal/storage"
	"fmt"
	"time"

	"log/slog"
)

func Daemon(ctx context.Context, cfg app.Config, redisClient *storage.Redis) error {
	slog.Info("starting process")
	defer slog.Info("stopping process")

	// Setup the spur api client
	spurAPI := spur.NewAPI("https://feeds.spur.us", "v2", cfg.SpurAPIToken)
	slog.Info(
		"spur api client created",
		slog.String("api_version", spurAPI.Version),
		slog.String("base_url", spurAPI.BaseURL),
	)

	// check redis for the latest feed info, in case we restarted
	lastFeedInfo, err := redisClient.GetLatestFeedInfo(ctx)
	if err != nil {
		lastFeedInfo = &spur.FeedInfo{}
	}

	// check redis for the latest merged data, in case we restarted
	lastRealtimeInfo, err := redisClient.GetLatestRealtimeFeedInfo(ctx)
	if err != nil {
		lastRealtimeInfo = &spur.RealtimeFeedInfo{}
	}

	// If we don't have the latest feed info, download and process the latest feed file to seed the initial data
	if lastFeedInfo.JSON.Date == "" {
		lastFeedInfo, err = spurAPI.LatestFeedInfo(ctx, cfg.SpurFeedType)
		if err != nil {
			return fmt.Errorf("error getting latest feed info: %v", err)
		}

		slog.Info("no initial data found, downloading latest feed")
		feedStream, err := spurAPI.LatestFeed(ctx, cfg.SpurFeedType)
		if err != nil {
			return fmt.Errorf("error getting latest feed: %v", err)
		}

		count, err := redisClient.StreamingFeedInsert(ctx, feedStream)
		if err != nil {
			return fmt.Errorf("error inserting feed into redis: %v", err)
		}

		err = redisClient.PutLatestFeedInfo(ctx, lastFeedInfo)
		if err != nil {
			return fmt.Errorf("error storing latest feed info: %v", err)
		}

		slog.Info("feed inserted into redis", slog.Int64("count", count))

		// Reprocess all the realtime data from the feed date 00:00:00 until now
		if cfg.SpurRealtimeEnabled {
			err := reprocessRealtime(ctx, redisClient, spurAPI, lastFeedInfo.JSON.Date)
			if err != nil {
				return fmt.Errorf("error reprocessing realtime data: %v", err)
			}
		}
	}

	// check for new data every minute
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	// Processing loop, check for new data every minute. If the feed info has changed, download the new data.
	// If the realtime info has changed, merge in the new data.
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			slog.Info("checking for new full feed data")
			latestFeedInfo, err := spurAPI.LatestFeedInfo(ctx, cfg.SpurFeedType)
			if err != nil {
				slog.Error("error getting latest feed info", "error", err.Error())
				continue
			}

			// If the feed info has changed, get the new data
			if latestFeedInfo.JSON.Date != lastFeedInfo.JSON.Date {
				err := processLatestFeedFile(ctx, latestFeedInfo, redisClient, spurAPI)
				if err != nil {
					slog.Error("error processing latest feed file", "error", err.Error())
					continue
				}
				lastFeedInfo = latestFeedInfo

				// Reprocess all the realtime data from the feed date 00:00:00 until now
				if cfg.SpurRealtimeEnabled {
					err := reprocessRealtime(ctx, redisClient, spurAPI, latestFeedInfo.JSON.Date)
					if err != nil {
						slog.Error("error reprocessing realtime data", "error", err.Error())
					}
				}

				// Reprocessing will take care of getting the latest realtime info, so we can skip the rest of this loop
				continue
			}

			// Realtime must be enabled to process realtime data
			if !cfg.SpurRealtimeEnabled {
				continue
			}

			slog.Info("checking for new realtime feed data")
			latestRealtimeInfo, err := spurAPI.LatestRealtimeFeedInfo(ctx, spur.AnonymousResidential)
			if err != nil {
				slog.Error("error getting latest realtime feed info", "error", err.Error())
				continue
			}

			// If the realtime info has changed, merge in the new data
			if latestRealtimeInfo.JSON.Date != lastRealtimeInfo.JSON.Date {
				err := processLatestRealtimeFeedFile(ctx, latestRealtimeInfo, redisClient, spurAPI)
				if err != nil {
					slog.Error("error processing latest realtime feed file", "error", err.Error())
					continue
				}
				lastRealtimeInfo = latestRealtimeInfo
			}
		}
	}

}

// processLatestFeedFile - download and process the latest feed file
func processLatestFeedFile(ctx context.Context, latestFeedInfo *spur.FeedInfo, redisClient *storage.Redis, spurAPI *spur.API) error {
	slog.Info("new feed info found, downloading latest feed")

	// Now download the latest feed file and process it
	slog.Info("processing the latest feed file")
	feedStream, err := spurAPI.LatestFeed(ctx, spur.AnonymousResidential)
	if err != nil {
		return fmt.Errorf("error getting latest feed: %v", err)
	}

	// insert the feed into redis
	count, err := redisClient.StreamingFeedInsert(ctx, feedStream)
	if err != nil {
		return fmt.Errorf("error inserting feed into redis: %v", err)
	}

	// we are done so store the latest feed info to redis
	err = redisClient.PutLatestFeedInfo(ctx, latestFeedInfo)
	if err != nil {
		return fmt.Errorf("error storing latest feed info: %v", err)
	}

	slog.Info("feed inserted into redis", slog.Int64("count", count))

	return nil
}

// processLatestRealtimeFeedFile - download and process the latest realtime feed file
func processLatestRealtimeFeedFile(ctx context.Context, latestRealtimeInfo *spur.RealtimeFeedInfo, redisClient *storage.Redis, spurAPI *spur.API) error {
	slog.Info("new realtime feed info found, downloading latest realtime feed")

	// Now download the latest realtime feed file and process it
	slog.Info("processing the latest realtime feed file")
	realtimeFeedStream, err := spurAPI.LatestRealtimeFeed(ctx, spur.AnonymousResidential)
	if err != nil {
		return fmt.Errorf("error getting latest realtime feed: %v", err)
	}

	// insert the realtime feed into redis
	count, err := redisClient.StreamingMergeInsert(ctx, realtimeFeedStream)
	if err != nil {
		return fmt.Errorf("error inserting realtime feed into redis: %v", err)
	}

	// we are done so store the latest feed info to redis
	err = redisClient.PutLatestRealtimeFeedInfo(ctx, latestRealtimeInfo)
	if err != nil {
		return fmt.Errorf("error storing latest realtime feed info: %v", err)
	}

	slog.Info("realtime feed merged into redis", slog.Int64("count", count))

	return nil
}

// reprocessRealtime - reprocess all realtime data from the given feed date until now
func reprocessRealtime(ctx context.Context, redisClient *storage.Redis, spurAPI *spur.API, feedDate string) error {
	latestFeedDate, err := time.Parse("20060102", feedDate)
	if err != nil {
		return fmt.Errorf("error parsing latest feed date: %v", err)
	}

	// Make sure the latest feed date is in UTC
	latestFeedDate = latestFeedDate.UTC()

	// Starting from latest feed date pull all realtime data until now, incrementing by 5 minutes each time
	currentTime := latestFeedDate
	totalCount := int64(0)
	for currentTime.Before(time.Now().UTC()) {
		slog.Info("processing realtime file for time", "time", currentTime.Format(time.RFC3339))
		realtimeFeedStream, err := spurAPI.RealtimeFeed(ctx, spur.AnonymousResidential, currentTime)
		if err != nil {
			return fmt.Errorf("error getting realtime feed: %v", err)
		}

		count, err := redisClient.StreamingMergeInsert(ctx, realtimeFeedStream)
		if err != nil {
			return fmt.Errorf("error inserting realtime feed into redis: %v", err)
		}

		currentTime = currentTime.Add(5 * time.Minute)
		totalCount += count
	}

	slog.Info("reprocessed historical realtime feed into redis", slog.Int64("count", totalCount))

	return nil
}
