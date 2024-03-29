package app

import (
	"feedexampleredis/internal/spur"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
)

// Config - the configuration for the process, parsed from environment variables
type Config struct {
	ChunkSize           int
	TTL                 int
	RedisAddr           string
	RedisPass           string
	RedisDB             int
	ConcurrentNum       int
	SpurAPIToken        string
	SpurFeedType        spur.FeedType
	SpurRealtimeEnabled bool
	Port                int
	LocalAPIAuthTokens  []string
	CertFile            string
	KeyFile             string
}

// parseConfig - parse the configuration from environment variables
func ParseConfigFromEnvironment() (Config, error) {
	cfg := Config{
		ChunkSize:           5000,
		TTL:                 24,
		RedisAddr:           "localhost:6379",
		RedisPass:           "",
		RedisDB:             0,
		ConcurrentNum:       runtime.NumCPU(),
		SpurAPIToken:        "",
		SpurFeedType:        spur.AnonymousFeed,
		SpurRealtimeEnabled: false,
		Port:                8080,
		LocalAPIAuthTokens:  nil,
		CertFile:            "",
		KeyFile:             "",
	}

	envChunkSize := os.Getenv("SPUR_REDIS_CHUNK_SIZE")
	if envChunkSize != "" {
		intChunkSize, err := strconv.Atoi(envChunkSize)
		if err != nil {
			return Config{}, fmt.Errorf("invalid SPUR_REDIS_CHUNK_SIZE: %v", err)
		}
		cfg.ChunkSize = intChunkSize
	}

	envTTL := os.Getenv("SPUR_REDIS_TTL")
	if envTTL != "" {
		intTTL, err := strconv.Atoi(envTTL)
		if err != nil {
			return Config{}, fmt.Errorf("invalid SPUR_REDIS_TTL: %v", err)
		}
		cfg.TTL = intTTL
	}

	envRedisAddr := os.Getenv("SPUR_REDIS_ADDR")
	if envRedisAddr != "" {
		cfg.RedisAddr = envRedisAddr
	}

	envRedisPass := os.Getenv("SPUR_REDIS_PASS")
	if envRedisPass != "" {
		cfg.RedisPass = envRedisPass
	}

	envRedisDB := os.Getenv("SPUR_REDIS_DB")
	if envRedisDB != "" {
		intRedisDB, err := strconv.Atoi(envRedisDB)
		if err != nil {
			return Config{}, fmt.Errorf("invalid SPUR_REDIS_DB: %v", err)
		}
		cfg.RedisDB = intRedisDB
	}

	envConcurrentNum := os.Getenv("SPUR_REDIS_CONCURRENT_NUM")
	if envConcurrentNum != "" {
		intConcurrentNum, err := strconv.Atoi(envConcurrentNum)
		if err != nil {
			return Config{}, fmt.Errorf("invalid SPUR_REDIS_CONCURRENT_NUM: %v", err)
		}

		if intConcurrentNum > runtime.NumCPU() {
			cfg.ConcurrentNum = runtime.NumCPU()
		} else {
			cfg.ConcurrentNum = intConcurrentNum
		}
	}

	envSpurAPIToken := os.Getenv("SPUR_REDIS_API_TOKEN")
	if envSpurAPIToken != "" {
		cfg.SpurAPIToken = envSpurAPIToken
	} else {
		return Config{}, fmt.Errorf("SPUR_REDIS_API_TOKEN is required")
	}

	envSpurFeedType := os.Getenv("SPUR_REDIS_FEED_TYPE")
	if envSpurFeedType != "" {
		cfg.SpurFeedType = spur.FeedType(envSpurFeedType)
	} else {
		cfg.SpurFeedType = spur.AnonymousFeed
	}

	envSpurRealtimeEnabled := os.Getenv("SPUR_REDIS_REALTIME_ENABLED")
	if envSpurRealtimeEnabled != "" {
		boolSpurRealtimeEnabled, err := strconv.ParseBool(envSpurRealtimeEnabled)
		if err != nil {
			return Config{}, fmt.Errorf("invalid SPUR_REDIS_REALTIME_ENABLED: %v", err)
		}
		cfg.SpurRealtimeEnabled = boolSpurRealtimeEnabled
	} else {
		cfg.SpurRealtimeEnabled = false

	}

	envPort := os.Getenv("SPUR_REDIS_PORT")
	if envPort != "" {
		intPort, err := strconv.Atoi(envPort)
		if err != nil {
			return Config{}, fmt.Errorf("invalid SPUR_REDIS_PORT: %v", err)
		}
		cfg.Port = intPort
	}

	envCertFile := os.Getenv("SPUR_REDIS_CERT_FILE")
	if envCertFile != "" {
		cfg.CertFile = envCertFile
	}

	envKeyFile := os.Getenv("SPUR_REDIS_KEY_FILE")
	if envKeyFile != "" {
		cfg.KeyFile = envKeyFile
	}

	envLocalAPIAuthTokens := os.Getenv("SPUR_REDIS_LOCAL_API_AUTH_TOKENS")
	if envLocalAPIAuthTokens != "" {
		// Tokens are comma separated
		parsed := strings.Split(envLocalAPIAuthTokens, ",")
		for _, token := range parsed {
			cfg.LocalAPIAuthTokens = append(cfg.LocalAPIAuthTokens, token)
		}
	} else {
		return Config{}, fmt.Errorf("SPUR_REDIS_LOCAL_API_AUTH_TOKENS is required")
	}

	return cfg, nil
}
