package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppName          string
	Env              string // dev|prod
	HTTPPort         int
	DBDSN            string
	InMemory         bool
	ReadTimeout      time.Duration
	WriteTimeout     time.Duration
	CORSOrigins      []string
	LogLevel         string
	PprofEnabled     bool
	APIKeys          []string
	CacheMaxCost     int64
	CacheNumCounters int64
	CacheBufferItems int64
	RedisAddr        string
	RedisPassword    string
	RedisDB          int
}

func Load() (*Config, error) {
	cfg := &Config{}
	cfg.AppName = getEnvDefault("APP_NAME", "maxwell-api")
	cfg.Env = getEnvDefault("ENV", "dev")
	cfg.DBDSN = os.Getenv("DB_DSN")
	// In-memory mode for local manual testing (skip external Postgres)
	if v := os.Getenv("INMEMORY"); v == "1" || strings.ToLower(v) == "true" {
		cfg.InMemory = true
	}
	portStr := getEnvDefault("HTTP_PORT", "8080")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid HTTP_PORT: %w", err)
	}
	cfg.HTTPPort = port

	read := getEnvDefault("READ_TIMEOUT", "10s")
	rt, err := time.ParseDuration(read)
	if err != nil {
		return nil, fmt.Errorf("invalid READ_TIMEOUT: %w", err)
	}
	cfg.ReadTimeout = rt
	write := getEnvDefault("WRITE_TIMEOUT", "15s")
	wt, err := time.ParseDuration(write)
	if err != nil {
		return nil, fmt.Errorf("invalid WRITE_TIMEOUT: %w", err)
	}
	cfg.WriteTimeout = wt

	origins := os.Getenv("CORS_ORIGINS")
	if origins != "" {
		cfg.CORSOrigins = strings.Split(origins, ",")
	}
	if v := os.Getenv("API_KEYS"); v != "" {
		parts := strings.Split(v, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				cfg.APIKeys = append(cfg.APIKeys, p)
			}
		}
	}
	cfg.LogLevel = getEnvDefault("LOG_LEVEL", "info")
	cfg.PprofEnabled = os.Getenv("PPROF_ENABLED") == "1"

	// Cache defaults
	cfg.CacheMaxCost = parseInt64Env("CACHE_MAX_COST", 10_000)
	cfg.CacheNumCounters = parseInt64Env("CACHE_NUM_COUNTERS", 100_000)
	cfg.CacheBufferItems = parseInt64Env("CACHE_BUFFER_ITEMS", 64)
	cfg.RedisAddr = getEnvDefault("REDIS_ADDR", "")
	cfg.RedisPassword = getEnvDefault("REDIS_PASSWORD", "")
	if v := os.Getenv("REDIS_DB"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.RedisDB = i
		}
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) Validate() error {
	if !c.InMemory && c.DBDSN == "" {
		return errors.New("DB_DSN required (or set INMEMORY=1 for in-memory store)")
	}
	if c.Env != "dev" && c.Env != "prod" {
		return fmt.Errorf("ENV must be dev or prod")
	}
	return nil
}

func getEnvDefault(k, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}

func parseInt64Env(k string, def int64) int64 {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	i, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return def
	}
	return i
}
