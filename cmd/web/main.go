// Command web serves the blog and portfolio.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mattbx/mattbx-go/internal/config"
	"github.com/mattbx/mattbx-go/internal/db"
	"github.com/mattbx/mattbx-go/internal/handlers"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	if err := run(log); err != nil {
		log.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run(log *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	sqlDB, err := db.Open(cfg.DBPath)
	if err != nil {
		return err
	}
	defer sqlDB.Close()
	log.Info("database ready", "path", cfg.DBPath)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handlers.New(cfg, sqlDB, log).Routes(),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// Disco sends SIGTERM during a deploy and waits before killing the
	// container, so drain in-flight requests instead of dropping them.
	shutdownDone := make(chan struct{})
	go func() {
		defer close(shutdownDone)

		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		log.Info("shutting down")
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Error("graceful shutdown failed", "err", err)
		}
	}()

	log.Info("listening", "addr", srv.Addr, "env", cfg.Env, "base_url", cfg.BaseURL)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	<-shutdownDone
	log.Info("stopped")
	return nil
}
