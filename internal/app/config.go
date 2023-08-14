package app

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
)

// Config - the configuration for the process, parsed from environment variables
type Config struct {
	ChunkSize     int
	TTL           int
	RedisAddr     string
	RedisPass     string
	RedisDB       int
	ConcurrentNum int
	SpurAPIToken  string
}

// parseConfig - parse the configuration from environment variables
func ParseConfigFromEnvironment() (Config, error) {
	cfg := Config{
		ChunkSize:     5000,
		TTL:           24,
		RedisAddr:     "localhost:6379",
		RedisPass:     "",
		RedisDB:       0,
		ConcurrentNum: runtime.NumCPU(),
		SpurAPIToken:  "",
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

	return cfg, nil
}
