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

	"github.com/chrisbirster/battlebunnywealth/internal/httpserver"
	"github.com/chrisbirster/battlebunnywealth/internal/proofofplay"
	webapp "github.com/chrisbirster/battlebunnywealth/web"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	addr := envOr("BBWEALTH_ADDR", ":8080")

	chain := proofofplay.NewChain(time.Unix(0, 0).UTC())
	protocol := proofofplay.NewProtocol(proofofplay.DefaultConfig(), chain)
	spa, err := webapp.Handler()
	if err != nil {
		logger.Error("create spa handler", "error", err)
		os.Exit(1)
	}

	handler := httpserver.New(logger, protocol, spa)
	server := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("shutdown server", "error", err)
		}
	}()

	logger.Info("battle bunny wealth listening", "addr", addr)
	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
