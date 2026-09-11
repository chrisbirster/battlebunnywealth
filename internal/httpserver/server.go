package httpserver

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/chrisbirster/battlebunnywealth/internal/game"
	"github.com/chrisbirster/battlebunnywealth/internal/proofofplay"
)

type proofOfPlay interface { Status(time.Time) proofofplay.NetworkStatus; Head() proofofplay.Block }

func New(logger *slog.Logger, protocol proofOfPlay, gameService *game.Service, spa http.Handler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/healthz", func(w http.ResponseWriter, _ *http.Request) { writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "battle-bunny-wealth", "version": "dev"}) })
	mux.HandleFunc("GET /api/v1/proof-of-play", func(w http.ResponseWriter, _ *http.Request) { writeJSON(w, http.StatusOK, protocol.Status(time.Now().UTC())) })
	mux.HandleFunc("GET /api/v1/proof-of-play/blocks/head", func(w http.ResponseWriter, _ *http.Request) { writeJSON(w, http.StatusOK, protocol.Head()) })
	mux.HandleFunc("GET /api/v1/game/state", func(w http.ResponseWriter, _ *http.Request) { snapshot, err := gameService.Snapshot(); if err != nil { writeGameError(w, err); return }; writeJSON(w, http.StatusOK, snapshot) })
	mux.HandleFunc("PUT /api/v1/game/profile", func(w http.ResponseWriter, r *http.Request) { var profile game.PlayerProfile; if err := decodeJSON(w, r, &profile); err != nil { writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()}); return }; snapshot, err := gameService.UpdateProfile(profile); if err != nil { writeGameError(w, err); return }; writeJSON(w, http.StatusOK, snapshot) })
	mux.HandleFunc("POST /api/v1/game/businesses/{id}/upgrade", func(w http.ResponseWriter, r *http.Request) { snapshot, err := gameService.UpgradeBusiness(r.PathValue("id")); if err != nil { writeGameError(w, err); return }; writeJSON(w, http.StatusOK, snapshot) })
	mux.HandleFunc("POST /api/v1/game/onboarding/advance", func(w http.ResponseWriter, _ *http.Request) { snapshot, err := gameService.AdvanceOnboarding(); if err != nil { writeGameError(w, err); return }; writeJSON(w, http.StatusOK, snapshot) })
	mux.HandleFunc("POST /api/v1/game/story/advance", func(w http.ResponseWriter, _ *http.Request) { snapshot, err := gameService.AdvanceStory(); if err != nil { writeGameError(w, err); return }; writeJSON(w, http.StatusOK, snapshot) })
	mux.HandleFunc("PUT /api/v1/game/warren", func(w http.ResponseWriter, r *http.Request) { var body struct { Theme string `json:"theme"` }; if err := decodeJSON(w, r, &body); err != nil { writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()}); return }; snapshot, err := gameService.UpdateWarren(body.Theme); if err != nil { writeGameError(w, err); return }; writeJSON(w, http.StatusOK, snapshot) })
	mux.HandleFunc("POST /api/v1/game/season/turn-in", func(w http.ResponseWriter, _ *http.Request) { snapshot, err := gameService.TurnInSeason(); if err != nil { writeGameError(w, err); return }; writeJSON(w, http.StatusOK, snapshot) })
	mux.HandleFunc("GET /api/v1/game/standings", func(w http.ResponseWriter, _ *http.Request) { standings, err := gameService.Standings(); if err != nil { writeGameError(w, err); return }; writeJSON(w, http.StatusOK, standings) })
	mux.HandleFunc("POST /api/v1/game/dev/season/advance", func(w http.ResponseWriter, _ *http.Request) { if os.Getenv("BBWEALTH_DEV_CONTROLS") != "1" { http.Error(w, "not found", http.StatusNotFound); return }; snapshot, err := gameService.AdvanceSeasonForDevelopment(); if err != nil { writeGameError(w, err); return }; writeJSON(w, http.StatusOK, snapshot) })
	mux.Handle("/", spa)
	return requestLog(logger, securityHeaders(mux))
}

func decodeJSON(w http.ResponseWriter, r *http.Request, value any) error { r.Body = http.MaxBytesReader(w, r.Body, 64<<10); decoder := json.NewDecoder(r.Body); decoder.DisallowUnknownFields(); return decoder.Decode(value) }
func writeGameError(w http.ResponseWriter, err error) { status := http.StatusInternalServerError; switch { case errors.Is(err, game.ErrInvalidProfile), errors.Is(err, game.ErrInvalidWarren): status = http.StatusBadRequest; case errors.Is(err, game.ErrUnknownBusiness): status = http.StatusNotFound; case errors.Is(err, game.ErrInsufficientFunds), errors.Is(err, game.ErrTurnInLocked), errors.Is(err, game.ErrStoryComplete): status = http.StatusConflict }; message := err.Error(); if status == http.StatusInternalServerError { message = "internal server error" }; writeJSON(w, status, map[string]string{"error": message}) }
func writeJSON(w http.ResponseWriter, status int, value any) { w.Header().Set("Content-Type", "application/json; charset=utf-8"); w.WriteHeader(status); _ = json.NewEncoder(w).Encode(value) }
func securityHeaders(next http.Handler) http.Handler { return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Header().Set("X-Content-Type-Options", "nosniff"); w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin"); w.Header().Set("X-Frame-Options", "DENY"); next.ServeHTTP(w, r) }) }
func requestLog(logger *slog.Logger, next http.Handler) http.Handler { return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { started := time.Now(); next.ServeHTTP(w, r); logger.Info("http request", "method", r.Method, "path", r.URL.Path, "duration", time.Since(started)) }) }
