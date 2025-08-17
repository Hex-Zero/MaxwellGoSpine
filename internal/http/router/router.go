package router

import (
    "database/sql"
    "net/http"
    "time"
    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    "go.uber.org/zap"
    appmw "github.com/hex-zero/MaxwellGoSpine/internal/middleware"
    "github.com/hex-zero/MaxwellGoSpine/internal/http/handlers"
    "github.com/hex-zero/MaxwellGoSpine/internal/http/render"
    "github.com/hex-zero/MaxwellGoSpine/internal/config"
    "github.com/hex-zero/MaxwellGoSpine/internal/core"
    "github.com/hex-zero/MaxwellGoSpine/internal/metrics"
    "net/http/pprof"
)

type Deps struct {
    Logger *zap.Logger
    UserSvc core.UserService
    CFG *config.Config
    Registry *metrics.Registry
    Version string
    Commit string
    BuildDate string
    DB *sql.DB
}

func New(d Deps) http.Handler {
    r := chi.NewRouter()
    r.Use(appmw.RequestID)
    r.Use(appmw.Recovery(d.Logger))
    r.Use(appmw.Logging(d.Logger, d.Registry))
    r.Use(appmw.Timeout(30 * time.Second))
    r.Use(appmw.CORS(d.CFG.CORSOrigins))
    r.Use(appmw.Gzip(-1))
    r.Use(appmw.ETag)
    r.Use(middleware.Heartbeat("/ping"))

    r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) { render.JSON(w, r, http.StatusOK, map[string]string{"status": "ok"}) })
    r.Get("/readyz", func(w http.ResponseWriter, r *http.Request) { if err := d.DB.PingContext(r.Context()); err != nil { http.Error(w, "db not ready", http.StatusServiceUnavailable); return }; render.JSON(w, r, http.StatusOK, map[string]string{"status": "ready"}) })
    r.Get("/version", func(w http.ResponseWriter, r *http.Request) { render.JSON(w, r, http.StatusOK, map[string]string{"version": d.Version, "commit": d.Commit, "date": d.BuildDate}) })

    if d.CFG.PprofEnabled {
        r.Mount("/debug/pprof", pprofHandler())
    }

    r.Route("/v1", func(api chi.Router) {
        handlers.NewUserHandler(d.UserSvc).Register(api)
    })
    return r
}

func pprofHandler() http.Handler {
    mux := chi.NewRouter()
    mux.Get("/", http.HandlerFunc(pprof.Index))
    mux.Get("/cmdline", http.HandlerFunc(pprof.Cmdline))
    mux.Get("/profile", http.HandlerFunc(pprof.Profile))
    mux.Get("/symbol", http.HandlerFunc(pprof.Symbol))
    mux.Get("/trace", http.HandlerFunc(pprof.Trace))
    return mux
}
