package httpserver

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/chrisbirster/battlebunnywealth/internal/proofofplay"
)

type proofOfPlay interface {
	Status(time.Time) proofofplay.NetworkStatus
	Head() proofofplay.Block
}

func New(logger *slog.Logger, protocol proofOfPlay, spa http.Handler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/v1/healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{
			"status":  "ok",
			"service": "battle-bunny-wealth",
			"version": "dev",
		})
	})

	mux.HandleFunc("GET /api/v1/proof-of-play", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, protocol.Status(time.Now().UTC()))
	})

	mux.HandleFunc("GET /api/v1/proof-of-play/blocks/head", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, protocol.Head())
	})

	mux.Handle("/", spa)
	return requestLog(logger, securityHeaders(mux))
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("X-Frame-Options", "DENY")
		next.ServeHTTP(w, r)
	})
}

func requestLog(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		next.ServeHTTP(w, r)
		logger.Info("http request", "method", r.Method, "path", r.URL.Path, "duration", time.Since(started))
	})
}
