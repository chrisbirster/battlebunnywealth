package httpserver

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/chrisbirster/battlebunnywealth/internal/game"
	"github.com/chrisbirster/battlebunnywealth/internal/proofofplay"
)

func testHandler(t *testing.T) http.Handler {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	protocol := proofofplay.NewProtocol(proofofplay.DefaultConfig(), proofofplay.NewChain(time.Unix(0, 0)))
	gameService, err := game.NewService(func() time.Time { return time.Date(2026, 9, 11, 12, 0, 0, 0, time.UTC) }, game.NewMemoryStore())
	if err != nil { t.Fatal(err) }
	return New(logger, protocol, gameService, http.NotFoundHandler())
}

func TestHealth(t *testing.T) {
	handler := testHandler(t)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/v1/healthz", nil))
	if rr.Code != http.StatusOK { t.Fatalf("status=%d want %d", rr.Code, http.StatusOK) }
}

func TestGameStateAndProfile(t *testing.T) {
	handler := testHandler(t)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/v1/game/state", nil))
	if rr.Code != http.StatusOK { t.Fatalf("state status=%d", rr.Code) }

	body := bytes.NewBufferString(`{"name":"Boomtail","callsign":"Fuse","fur":"charcoal","ears":"battle-worn","uniform":"night-black","cosmetics":[]}`)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodPut, "/api/v1/game/profile", body))
	if rr.Code != http.StatusOK { t.Fatalf("profile status=%d body=%s", rr.Code, rr.Body.String()) }
}
