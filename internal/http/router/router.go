package router

import (
	"database/sql"
	"io"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/hex-zero/MaxwellGoSpine/graph/resolver"
	"github.com/hex-zero/MaxwellGoSpine/graph/server"
	"github.com/hex-zero/MaxwellGoSpine/internal/config"
	"github.com/hex-zero/MaxwellGoSpine/internal/core"
	"github.com/hex-zero/MaxwellGoSpine/internal/http/handlers"
	"github.com/hex-zero/MaxwellGoSpine/internal/http/render"
	"github.com/hex-zero/MaxwellGoSpine/internal/metrics"
	appmw "github.com/hex-zero/MaxwellGoSpine/internal/middleware"
	"go.uber.org/zap"
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

	// Optional: GraphQL playground (gqlgen's) can be enabled in dev only.
	if d.CFG.Env == "dev" {
		r.Get("/playground", func(w http.ResponseWriter, r *http.Request) {
			server.PlaygroundHandler().ServeHTTP(w, r)
		})
	}

	r.Route("/v1", func(api chi.Router) {
		// Secure all versioned API endpoints with API key if configured
		// Build expiries map (per-key start-of-day Unix seconds) for middleware
		expUnix := map[string]int64{}
		for k, t := range d.CFG.APIKeyExpiries {
			// floor to day boundary UTC
			day := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
			expUnix[k] = day.Unix()
		}
		api.Use(appmw.APIKeyAuthWithOpts(appmw.APIKeyOptions{Current: d.CFG.APIKeys, Old: d.CFG.OldAPIKeys, Expiries: expUnix}))
		// REST handlers
		handlers.NewUserHandler(d.UserSvc).Register(api)
		// GraphQL endpoint (gqlgen executable schema)
		resolvers := &resolver.Resolver{UserService: d.UserSvc}
		gqlServer := server.NewExecutableSchema(resolvers)
		api.Handle("/graphql", gqlServer)
	})

	// Note: custom GraphiQL UI removed; use /playground (dev only) or external client.

	return r
}

// redocHTML serves the ReDoc UI for the OpenAPI spec.
const redocHTML = `<!DOCTYPE html>
	<html>
	<head>
		<meta charset="utf-8" />
		<title>API Docs</title>
		<link rel="shortcut icon" href="https://ReDoc.ly/favicon.ico" />
		<style>body { margin: 0; padding: 0; } </style>
	</head>
	<body>
		<redoc spec-url='/openapi.yaml'></redoc>
		<script src="https://cdn.redoc.ly/redoc/latest/bundles/redoc.standalone.js"></script>
	</body>
	</html>`

// pprofHandler mounts the Go pprof handlers under /debug/pprof.
func pprofHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", pprof.Index)
	mux.HandleFunc("/cmdline", pprof.Cmdline)
	mux.HandleFunc("/profile", pprof.Profile)
	mux.HandleFunc("/symbol", pprof.Symbol)
	mux.HandleFunc("/trace", pprof.Trace)
	return mux
}
