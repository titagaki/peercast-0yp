package main

import (
	"context"
	"database/sql"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/megan/peercast-root-shim/archive"
	"github.com/megan/peercast-root-shim/channel"
	"github.com/megan/peercast-root-shim/config"
	"github.com/megan/peercast-root-shim/httpd"
	"github.com/megan/peercast-root-shim/server"
)

func main() {
	configPath := flag.String("config", "./peercast-root-shim.toml", "path to config file")
	flag.Parse()

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

	srv, err := server.New(store)
	if err != nil {
		slog.Error("failed to create PCP server", "err", err)
		os.Exit(1)
	}

	rec := archive.New(db, store, slog.Default())

	httpdSrv := httpd.New(httpd.Config{
		Addr:        cfg.HTTP.Addr,
		CORSOrigins: cfg.HTTP.CORSOrigins,
	}, store, db)

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

	slog.Info("starting", "pcp", cfg.PCP.Addr, "http", cfg.HTTP.Addr)
	if err := srv.ListenAndServe(ctx, cfg.PCP.Addr); err != nil {
		slog.Error("PCP server stopped with error", "err", err)
		os.Exit(1)
	}

	wg.Wait()
	slog.Info("server stopped")
}
