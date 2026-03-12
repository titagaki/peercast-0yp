// Package httpd provides the HTTP server for the PeerCast YP.
package httpd

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/titagaki/peercast-0yp/internal/channel"
	"github.com/titagaki/peercast-0yp/internal/repository"
)

// InfoLine is a single announcement entry shown in index.txt.
type InfoLine struct {
	Name    string
	Comment string
	URL     string
}

// Server is the HTTP server for the PeerCast YP.
type Server struct {
	store      *channel.Store
	sessions   *repository.SessionRepo
	snapshots  *repository.SnapshotRepo
	router     chi.Router
	srv        *http.Server
	ypName     string
	ypURL      string
	ypIndexURL string
	pcpAddress string
	startTime  time.Time
	infoLines  []InfoLine
}

// Config holds HTTP server configuration.
type Config struct {
	Port        int
	CORSOrigins []string
	YPName      string
	YPURL       string
	YPIndexURL  string
	PCPAddress  string
	InfoLines   []InfoLine
}

// New creates a Server and registers all routes.
func New(cfg Config, store *channel.Store, sessions *repository.SessionRepo, snapshots *repository.SnapshotRepo) *Server {
	s := &Server{
		store:      store,
		sessions:   sessions,
		snapshots:  snapshots,
		ypName:     cfg.YPName,
		ypURL:      cfg.YPURL,
		ypIndexURL: cfg.YPIndexURL,
		pcpAddress: cfg.PCPAddress,
		startTime:  time.Now(),
		infoLines:  cfg.InfoLines,
	}
	s.router = s.buildRouter(cfg.CORSOrigins)
	s.srv = &http.Server{Addr: fmt.Sprintf(":%d", cfg.Port), Handler: s.router}
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

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-8")
		w.Write(topPageHTML)
	})
	r.Get("/yp", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/yp/", http.StatusMovedPermanently)
	})
	r.Get("/yp/api/config", s.handleAPIConfig)
	r.Get("/yp/api/channels", s.handleAPIChannels)
	r.Get("/yp/api/channels/activity", s.handleAPIActivity)
	r.Get("/yp/api/channels/timeline", s.handleAPITimeline)
	r.Get("/yp/api/history", s.handleAPIHistory)
	r.Get("/yp/index.txt", s.handleIndexTxt)

	r.Handle("/yp/*", http.StripPrefix("/yp", spaHandler()))

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
