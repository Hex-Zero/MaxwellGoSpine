package main

import (
    "context"
    "errors"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/go-chi/chi/v5"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "go.uber.org/zap"

    "github.com/hex-zero/MaxwellGoSpine/internal/config"
    "github.com/hex-zero/MaxwellGoSpine/internal/core"
    routerpkg "github.com/hex-zero/MaxwellGoSpine/internal/http/router"
    applog "github.com/hex-zero/MaxwellGoSpine/internal/log"
    "github.com/hex-zero/MaxwellGoSpine/internal/metrics"
    "github.com/hex-zero/MaxwellGoSpine/internal/storage/postgres"
)

var (
    version = "dev"
    commit  = "none"
    date    = "unknown"
)

func main() {
    ctx := context.Background()
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("config error: %v", err)
    }

    logger, err := applog.New(cfg.Env, cfg.LogLevel)
    if err != nil {
        log.Fatalf("logger init error: %v", err)
    }
    defer logger.Sync() //nolint:errcheck

    logger.Info("starting server", zap.String("version", version), zap.String("commit", commit), zap.String("date", date))

    db, err := postgres.Open(ctx, cfg.DBDSN)
    if err != nil {
        logger.Fatal("db open", zap.Error(err))
    }
    defer db.Close()

    userRepo := postgres.NewUserRepo(db)
    userSvc := core.NewUserService(userRepo)

    reg := metrics.NewRegistry()

    r := chi.NewRouter()
    apiRouter := routerpkg.New(routerpkg.Deps{
        Logger:   logger,
        UserSvc:  userSvc,
        CFG:      cfg,
        Registry: reg,
        Version:  version,
        Commit:   commit,
        BuildDate: date,
        DB:       db,
    })

    r.Mount("/", apiRouter)
    // metrics endpoint
    r.Handle("/metrics", promhttp.HandlerFor(reg.Gatherer, promhttp.HandlerOpts{}))

    srv := &http.Server{
        Addr:              fmt.Sprintf(":%d", cfg.HTTPPort),
        Handler:           r,
        ReadTimeout:       cfg.ReadTimeout,
        ReadHeaderTimeout: 5 * time.Second,
        WriteTimeout:      cfg.WriteTimeout,
        IdleTimeout:       60 * time.Second,
        ErrorLog:          zap.NewStdLog(logger.Named("http_error")),
    }

    go func() {
        logger.Info("http server listening", zap.Int("port", cfg.HTTPPort))
        if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
            logger.Fatal("server error", zap.Error(err))
        }
    }()

    stop := make(chan os.Signal, 1)
    signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
    <-stop

    shutdownCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
    defer cancel()
    if err := srv.Shutdown(shutdownCtx); err != nil {
        logger.Error("graceful shutdown failed", zap.Error(err))
        _ = srv.Close()
    }
    logger.Info("server stopped")
}
