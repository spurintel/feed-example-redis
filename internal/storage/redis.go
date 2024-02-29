package storage

import (
	"bufio"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"feedexampleredis/internal/spur"

	"github.com/go-redis/redis/v8"
	jsoniter "github.com/json-iterator/go"
)

type Redis struct {
	addr        string
	password    string
	db          int
	ttl         time.Duration
	concurrency int
	chunkSize   int
	client      *redis.Client
}

// NewRedis - create a new Redis storage object
func NewRedis(addr, password string, db int, ttl time.Duration, concurrency int, chunkSize int) *Redis {
	return &Redis{
		addr:        addr,
		password:    password,
		db:          db,
		ttl:         ttl,
		concurrency: concurrency,
		chunkSize:   chunkSize,
	}
}

// Connect - connect to the Redis server
func (r *Redis) Connect() error {
	r.client = redis.NewClient(&redis.Options{
		Addr:     r.addr,
		Password: r.password,
		DB:       r.db,
	})
	return nil
}

// Close - close the connection to the Redis server
func (r *Redis) Close() error {
	return r.client.Close()
}

// GetByIP - get an IP context from Redis where the IP is the key
func (r *Redis) GetByIP(ctx context.Context, ip string) (*spur.IPContext, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	val, err := r.client.Get(ctx, ip).Result()
	if err != nil {
		return nil, err
	}

	var ipctx spur.IPContext
	err = json.Unmarshal([]byte(val), &ipctx)
	if err != nil {
		return nil, err
	}

	return &ipctx, nil
}

// LatestFeedInfo - get the latest feed info from Redis
func (r *Redis) GetLatestFeedInfo(ctx context.Context) (*spur.FeedInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	val, err := r.client.Get(ctx, "feed_info").Result()
	if err != nil {
		return nil, err
	}

	var fi spur.FeedInfo
	err = json.Unmarshal([]byte(val), &fi)
	if err != nil {
		return nil, err
	}

	return &fi, nil
}

// PutLatestFeedInfo - put the latest feed info into Redis
func (r *Redis) PutLatestFeedInfo(ctx context.Context, fi *spur.FeedInfo) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	val, err := json.Marshal(fi)
	if err != nil {
		return err
	}

	err = r.client.Set(ctx, "feed_info", val, 0).Err()
	if err != nil {
		return err
	}

	return nil
}

// GetLatestRealtimeFeedInfo - get the latest realtime feed info from Redis
func (r *Redis) GetLatestRealtimeFeedInfo(ctx context.Context) (*spur.RealtimeFeedInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	val, err := r.client.Get(ctx, "realtime_feed_info").Result()
	if err != nil {
		return nil, err
	}

	var fi spur.RealtimeFeedInfo
	err = json.Unmarshal([]byte(val), &fi)
	if err != nil {
		return nil, err
	}

	return &fi, nil
}

// PutLatestRealtimeFeedInfo - put the latest realtime feed info into Redis
func (r *Redis) PutLatestRealtimeFeedInfo(ctx context.Context, fi *spur.RealtimeFeedInfo) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	val, err := json.Marshal(fi)
	if err != nil {
		return err
	}

	err = r.client.Set(ctx, "realtime_feed_info", val, 0).Err()
	if err != nil {
		return err
	}

	return nil
}

// StreamingFeedInsert - insert a streaming feed file download into Redis using a pipeline. This will overwrite any existing keys with the new data.
func (r *Redis) StreamingFeedInsert(ctx context.Context, rc io.ReadCloser) (int64, error) {
	defer rc.Close()

	// Feed donwloads are gzipped, so we need to decompress them
	gzr, err := gzip.NewReader(rc)
	if err != nil {
		return 0, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	var wg sync.WaitGroup
	lines := readLines(ctx, gzr, r.concurrency, r.chunkSize)
	var count int64
	for i := 0; i < r.concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			ctx := context.Background()
			processed, err := processFeedLines(ctx, r.chunkSize, r.ttl, workerID, lines, r.client)
			if err != nil {
				slog.Error("failed to process feed lines", "error", err.Error())
				return
			}
			atomic.AddInt64(&count, processed)
		}(i)
	}

	wg.Wait()
	return count, nil
}

// StreamingMergeInsert - insert a streaming realtime update file download into Redis using a pipeline. This will merge data with existing keys.
func (r *Redis) StreamingMergeInsert(ctx context.Context, rc io.ReadCloser) (int64, error) {
	defer rc.Close()

	// Feed donwloads are gzipped, so we need to decompress them
	gzr, err := gzip.NewReader(rc)
	if err != nil {
		return 0, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	var wg sync.WaitGroup
	lines := readLines(ctx, gzr, r.concurrency, r.chunkSize)
	var count int64
	for i := 0; i < r.concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			ctx := context.Background()
			processed, err := processMergeLines(ctx, r.chunkSize, r.ttl, workerID, lines, r.client)
			if err != nil {
				slog.Error("failed to process feed lines", "error", err.Error())
				return
			}

			atomic.AddInt64(&count, processed)
		}(i)
	}

	wg.Wait()
	return count, nil
}

func readLines(ctx context.Context, r io.Reader, concurrency int, chunkSize int) <-chan []byte {
	lines := make(chan []byte, concurrency*chunkSize)
	go func() {
		defer close(lines)
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			if ctx.Err() != nil {
				return
			}

			line := scanner.Bytes()
			b := make([]byte, len(line))
			copy(b, line)
			lines <- b
		}
	}()
	return lines
}

func processFeedLines(ctx context.Context, chunkSize int, ttl time.Duration, workerID int, lines <-chan []byte, rdb *redis.Client) (int64, error) {
	pipe := rdb.Pipeline()
	buffer := 0
	count := int64(0)

	for line := range lines {
		var record spur.IPContext
		if err := json.Unmarshal(line, &record); err != nil {
			slog.Error("error unmarshalling line", "worker_id", workerID, "error", err.Error())
			continue
		}

		buffer++
		key := record.IP
		pipe.Set(ctx, key, string(line), time.Duration(ttl)*time.Hour)
		if buffer >= chunkSize {
			result, err := pipe.Exec(ctx)
			if err != nil {
				return 0, fmt.Errorf("worker %d: error executing pipeline: %w", workerID, err)
			}

			for _, res := range result {
				if res.Err() != nil {
					return 0, fmt.Errorf("worker %d: error executing pipeline: %w", workerID, err)
				}
			}

			count += int64(buffer)
			buffer = 0
		}
	}
	if buffer > 0 {
		// fmt.Printf("Worker %d: Flushing (%d)\n", workerID, count)
		_, err := pipe.Exec(ctx)
		if err != nil {
			return 0, fmt.Errorf("worker %d: error executing pipeline: %w", workerID, err)
		}

		count += int64(buffer)
	}

	return count, nil
}

func processMergeLines(ctx context.Context, chunkSize int, ttl time.Duration, workerID int, lines <-chan []byte, rdb *redis.Client) (int64, error) {
	pipe := rdb.Pipeline()
	buffer := 0
	count := int64(0)

	partials := make(map[string]*spur.IPContext)
	existing := make(map[string]*spur.IPContext)

	// Load all of our new lines into partials
	for line := range lines {
		var record spur.IPContext
		if err := jsoniter.Unmarshal(line, &record); err != nil {
			fmt.Printf("Worker %d: Skipping failed JSON: %s\n", workerID, line)
			continue
		}
		partials[record.IP] = &record
	}

	// Fetch all of the IPs we have updates for from Redis
	keys := make([]string, count)
	for k := range partials {
		keys = append(keys, k)
	}

	state, err := rdb.MGet(ctx, keys...).Result()
	if err != nil {
		return 0, fmt.Errorf("worker %d: error fetching keys: %w", workerID, err)
	}

	for _, val := range state {
		var existingRecord spur.IPContext
		if val == nil {
			continue
		}
		data, ok := val.(string)
		if !ok {
			slog.Error("failed to cast to string", "worker_id", workerID, "value", val)
			continue
		}
		if err := json.Unmarshal([]byte(data), &existingRecord); err != nil {
			slog.Error("failed to unmarshal json", "worker_id", workerID, "error", err.Error())
			continue
		}
		existing[existingRecord.IP] = &existingRecord
	}

	// Merge all of our Redis IPs with the partial IPs
	for ip, partial := range partials {
		if existing[ip] != nil {
			eip := existing[ip]
			eip.Merge(partial)
			partial = existing[ip]
		}
		buffer++
		key := partial.IP
		data, err := json.Marshal(partial)
		if err != nil {
			log.Fatalf("Worker %d: Failed to serialize json: %v\n", workerID, err)
			continue
		}
		pipe.Set(ctx, key, string(data), time.Duration(ttl)*time.Hour)
		if buffer >= chunkSize {
			// fmt.Printf("\r\nWorker %d: Flushing (%d)", workerID, count)
			_, err = pipe.Exec(ctx)
			if err != nil {
				log.Fatalf("Worker %d: Error executing pipeline: %v\n", workerID, err)
			}
			count += int64(buffer)
			buffer = 0
		}
	}
	if buffer > 0 {
		// fmt.Printf("Worker %d: Flushing (%d)\n", workerID, count)
		_, err = pipe.Exec(ctx)
		if err != nil {
			log.Fatalf("Worker %d: Error executing pipeline: %v\n", workerID, err)
		}

		count += int64(buffer)
	}

	return count, nil
}
