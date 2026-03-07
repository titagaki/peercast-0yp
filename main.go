package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/titagaki/peercast-0yp/internal/archive"
	"github.com/titagaki/peercast-0yp/internal/channel"
	"github.com/titagaki/peercast-0yp/internal/config"
	"github.com/titagaki/peercast-0yp/internal/httpd"
	"github.com/titagaki/peercast-0yp/internal/repository"
	"github.com/titagaki/peercast-0yp/internal/server"
)

// loadDotEnv reads key=value pairs from .env and sets them as environment
// variables if they are not already set. Silently ignores missing file.
func loadDotEnv(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		if os.Getenv(k) == "" {
			os.Setenv(k, v)
		}
	}
}

func main() {
	configPath := flag.String("config", "./peercast-0yp.toml", "path to config file")
	flag.Parse()

	loadDotEnv(".env")

	cfg, err := config.Load(*configPath)
	if err != nil {
		slog.Error("failed to load config", "err", err)
		os.Exit(1)
	}

	db, err := sql.Open("mysql", cfg.Database.DSN)
	if err != nil {
		slog.Error("failed to open database", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	store := channel.NewStore()

	srv, err := server.New(store, server.Config{
		MaxConnections:   cfg.PCP.MaxConnections,
		UpdateInterval:   time.Duration(cfg.PCP.UpdateInterval) * time.Second,
		HitTimeout:       time.Duration(cfg.PCP.HitTimeout) * time.Second,
		MinClientVersion: cfg.PCP.MinClientVersion,
	})
	if err != nil {
		slog.Error("failed to create PCP server", "err", err)
		os.Exit(1)
	}

	sessions := repository.NewSessionRepo(db)
	snapshots := repository.NewSnapshotRepo(db)

	rec := archive.New(sessions, snapshots, store, slog.Default())

	httpdSrv := httpd.New(httpd.Config{
		Port:        cfg.HTTP.Port,
		CORSOrigins: cfg.HTTP.CORSOrigins,
	}, store, sessions, snapshots)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		rec.Start(ctx)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := httpdSrv.ListenAndServe(); err != nil {
			slog.Info("HTTP server stopped", "err", err)
		}
	}()
	go func() {
		<-ctx.Done()
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		httpdSrv.Shutdown(shutCtx)
	}()

	pcpAddr := fmt.Sprintf(":%d", cfg.PCP.Port)
	slog.Info("starting", "pcp", pcpAddr, "http", fmt.Sprintf(":%d", cfg.HTTP.Port))
	if err := srv.ListenAndServe(ctx, pcpAddr); err != nil {
		slog.Error("PCP server stopped with error", "err", err)
		os.Exit(1)
	}

	wg.Wait()
	slog.Info("server stopped")
}
