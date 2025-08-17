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
    AppName     string
    Env         string // dev|prod
    HTTPPort    int
    DBDSN       string
    ReadTimeout time.Duration
    WriteTimeout time.Duration
    CORSOrigins []string
    LogLevel    string
    PprofEnabled bool
}

func Load() (*Config, error) {
    cfg := &Config{}
    cfg.AppName = getEnvDefault("APP_NAME", "maxwell-api")
    cfg.Env = getEnvDefault("ENV", "dev")
    cfg.DBDSN = os.Getenv("DB_DSN")
    portStr := getEnvDefault("HTTP_PORT", "8080")
    port, err := strconv.Atoi(portStr)
    if err != nil {
        return nil, fmt.Errorf("invalid HTTP_PORT: %w", err)
    }
    cfg.HTTPPort = port

    read := getEnvDefault("READ_TIMEOUT", "10s")
    rt, err := time.ParseDuration(read)
    if err != nil { return nil, fmt.Errorf("invalid READ_TIMEOUT: %w", err) }
    cfg.ReadTimeout = rt
    write := getEnvDefault("WRITE_TIMEOUT", "15s")
    wt, err := time.ParseDuration(write)
    if err != nil { return nil, fmt.Errorf("invalid WRITE_TIMEOUT: %w", err) }
    cfg.WriteTimeout = wt

    origins := os.Getenv("CORS_ORIGINS")
    if origins != "" { cfg.CORSOrigins = strings.Split(origins, ",") }
    cfg.LogLevel = getEnvDefault("LOG_LEVEL", "info")
    cfg.PprofEnabled = os.Getenv("PPROF_ENABLED") == "1"

    if err := cfg.Validate(); err != nil { return nil, err }
    return cfg, nil
}

func (c *Config) Validate() error {
    if c.DBDSN == "" { return errors.New("DB_DSN required") }
    if c.Env != "dev" && c.Env != "prod" { return fmt.Errorf("ENV must be dev or prod") }
    return nil
}

func getEnvDefault(k, def string) string { v := os.Getenv(k); if v == "" { return def }; return v }
