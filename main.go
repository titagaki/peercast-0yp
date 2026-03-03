package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/megan/peercast-root-shim/api"
	"github.com/megan/peercast-root-shim/channel"
	"github.com/megan/peercast-root-shim/server"
)

func main() {
	store := channel.NewStore()

	srv, err := server.New(store)
	if err != nil {
		slog.Error("failed to create server", "err", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// HTTP API server
	mux := http.NewServeMux()
	api.NewHandler(store).Register(mux)
	httpSrv := &http.Server{Addr: ":7145", Handler: mux}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server error", "err", err)
		}
	}()
	go func() {
		<-ctx.Done()
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		httpSrv.Shutdown(shutCtx)
	}()

	// PCP server (blocks until ctx is cancelled)
	pcpAddr := ":7144"
	slog.Info("PeerCast root server starting", "pcp", pcpAddr, "http", ":7145")
	if err := srv.ListenAndServe(ctx, pcpAddr); err != nil {
		slog.Error("PCP server stopped with error", "err", err)
		os.Exit(1)
	}

	wg.Wait()
	slog.Info("server stopped")
}
