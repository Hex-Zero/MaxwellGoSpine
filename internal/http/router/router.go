package router

import (
	"database/sql"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/hex-zero/MaxwellGoSpine/internal/config"
	"github.com/hex-zero/MaxwellGoSpine/internal/core"
	"github.com/hex-zero/MaxwellGoSpine/internal/http/handlers"
	"github.com/hex-zero/MaxwellGoSpine/internal/http/render"
	"github.com/hex-zero/MaxwellGoSpine/internal/metrics"
	appmw "github.com/hex-zero/MaxwellGoSpine/internal/middleware"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/http/pprof"
	"time"
)

type Deps struct {
	Logger    *zap.Logger
	UserSvc   core.UserService
	CFG       *config.Config
	Registry  *metrics.Registry
	Version   string
	Commit    string
	BuildDate string
	DB        *sql.DB
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

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		render.JSON(w, r, http.StatusOK, map[string]string{"status": "ok"})
	})
	r.Get("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if err := d.DB.PingContext(r.Context()); err != nil {
			http.Error(w, "db not ready", http.StatusServiceUnavailable)
			return
		}
		render.JSON(w, r, http.StatusOK, map[string]string{"status": "ready"})
	})
	r.Get("/version", func(w http.ResponseWriter, r *http.Request) {
		render.JSON(w, r, http.StatusOK, map[string]string{"version": d.Version, "commit": d.Commit, "date": d.BuildDate})
	})

	if d.CFG.PprofEnabled {
		r.Mount("/debug/pprof", pprofHandler())
	}

	// Serve raw OpenAPI spec
	r.Get("/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yaml")
		http.ServeFile(w, r, "openapi.yaml")
	})
	// ReDoc documentation UI
	r.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		io.WriteString(w, redocHTML)
	})

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

const redocHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="utf-8"/>
    <title>API Docs</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link rel="icon" href="data:,">
    <style>body{margin:0;padding:0;} .banner{position:fixed;top:0;left:0;right:0;background:#0d1117;color:#fff;font:14px/1.4 system-ui;padding:6px 12px;z-index:10} redoc{margin-top:32px}</style>
</head>
<body>
    <div class="banner">OpenAPI documentation - <a href="/openapi.yaml" style="color:#58a6ff">download spec</a></div>
    <redoc spec-url='/openapi.yaml'></redoc>
    <script src="https://cdn.redoc.ly/redoc/latest/bundles/redoc.standalone.js"></script>
</body>
</html>`
