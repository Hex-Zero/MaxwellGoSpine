package router

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/hex-zero/MaxwellGoSpine/internal/config"
	"github.com/hex-zero/MaxwellGoSpine/internal/core"
	gh "github.com/hex-zero/MaxwellGoSpine/internal/http/graphql"
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

	// Public GraphiQL playground (still sends API key for /v1/graphql requests)
	r.Get("/graphiql", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		io.WriteString(w, graphiqlHTML)
	})

	// Legacy path redirect kept as explicit route (no auth) so old bookmarks work.
	r.Get("/v1/graphiql", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/graphiql", http.StatusTemporaryRedirect)
	})

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
		// GraphQL endpoint (using graphql-go minimal implementation)
		schema, err := gh.BuildSchema(d.UserSvc)
		if err == nil {
			api.Post("/graphql", func(w http.ResponseWriter, r *http.Request) {
				var payload struct {
					Query     string         `json:"query"`
					Variables map[string]any `json:"variables"`
				}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "invalid body", http.StatusBadRequest)
					return
				}
				res := gh.Execute(r.Context(), schema, payload.Query, payload.Variables)
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(res)
			})
		}
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

const graphiqlHTML = `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="utf-8" />
	<title>GraphiQL</title>
	<meta name="viewport" content="width=device-width,initial-scale=1" />
	<link rel="stylesheet" href="https://unpkg.com/graphiql@3.1.0/graphiql.min.css" />
	<style>body,html{height:100%;margin:0} #graphiql{height:100vh;} .notice{position:absolute;top:6px;left:8px;z-index:10;background:#0d1117;color:#fff;font:12px system-ui;padding:4px 8px;border-radius:4px;opacity:.85}</style>
</head>
<body>
	<div class="notice">GraphQL Playground (public) – pass ?key=YOUR_API_KEY to override default devkey1</div>
	<div id="graphiql">Loading GraphiQL…</div>
	<script crossorigin src="https://unpkg.com/react@18.2.0/umd/react.production.min.js"></script>
	<script crossorigin src="https://unpkg.com/react-dom@18.2.0/umd/react-dom.production.min.js"></script>
	<script src="https://unpkg.com/graphiql@3.1.0/graphiql.min.js"></script>
	<script>
		(function(){
			function getParam(name){
				const m = new URLSearchParams(window.location.search).get(name); return m || '';
			}
			const apiKey = getParam('key') || 'devkey1';
			function buildHeaders(h){ return Object.assign({'X-API-Key': apiKey}, h||{}); }
			function makeFetcher(){
				if (window.GraphiQL && GraphiQL.createFetcher){
					return GraphiQL.createFetcher({ url: '/v1/graphql', headers: buildHeaders({'Content-Type':'application/json'}) });
				}
				// fallback manual fetcher
				return (params) => fetch('/v1/graphql', {method:'POST', headers: buildHeaders({'Content-Type':'application/json'}), body: JSON.stringify(params)}).then(r=>r.json());
			}
			function render(){
				if (!window.React || !window.ReactDOM) { return setTimeout(render, 30); }
				const fetcher = makeFetcher();
				const el = React.createElement(GraphiQL, { fetcher });
				if (ReactDOM.createRoot){
					ReactDOM.createRoot(document.getElementById('graphiql')).render(el);
				} else {
					ReactDOM.render(el, document.getElementById('graphiql'));
				}
			}
			render();
		})();
	</script>
</body>
</html>`
