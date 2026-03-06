// Package httpd provides the HTTP server for the PeerCast YP.
package httpd

import (
	"context"
	"database/sql"
	"embed"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/megan/peercast-root-shim/channel"
)

//go:embed static
var staticFiles embed.FS

// Server is the HTTP server for the PeerCast YP.
type Server struct {
	store  *channel.Store
	db     *sql.DB
	router chi.Router
	srv    *http.Server
}

// Config holds HTTP server configuration.
type Config struct {
	Addr        string
	CORSOrigins []string
}

// New creates a Server and registers all routes.
func New(cfg Config, store *channel.Store, db *sql.DB) *Server {
	s := &Server{store: store, db: db}
	s.router = s.buildRouter(cfg.CORSOrigins)
	s.srv = &http.Server{Addr: cfg.Addr, Handler: s.router}
	return s
}

func (s *Server) buildRouter(corsOrigins []string) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	if len(corsOrigins) > 0 {
		r.Use(cors.Handler(cors.Options{
			AllowedOrigins: corsOrigins,
			AllowedMethods: []string{"GET"},
			AllowedHeaders: []string{"Accept", "Content-Type"},
		}))
	}

	r.Get("/api/channels", s.handleAPIChannels)
	r.Get("/api/channels/{id}/activity", s.handleAPIActivity)
	r.Get("/api/channels/{id}/timeline", s.handleAPITimeline)
	r.Get("/api/history", s.handleAPIHistory)
	r.Get("/index.txt", s.handleIndexTxt)
	r.Get("/*", s.handleSPA)

	return r
}

// ListenAndServe starts the HTTP server.
func (s *Server) ListenAndServe() error {
	return s.srv.ListenAndServe()
}

// Shutdown gracefully stops the HTTP server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}
