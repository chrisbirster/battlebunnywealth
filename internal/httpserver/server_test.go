package httpserver

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/chrisbirster/battlebunnywealth/internal/game"
	"github.com/chrisbirster/battlebunnywealth/internal/identity"
	"github.com/chrisbirster/battlebunnywealth/internal/proofofplay"
)

func testHandler(t *testing.T) http.Handler {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	protocol := proofofplay.NewProtocol(proofofplay.DefaultConfig(), proofofplay.NewChain(time.Unix(0, 0)))
	now := func() time.Time { return time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC) }
	authority, err := proofofplay.NewAuthorityService(now, proofofplay.NewMemoryAuthorityStore(), proofofplay.DefaultAuthorityConfig())
	if err != nil { t.Fatal(err) }
	games := game.NewRegistry(now, filepath.Join(t.TempDir(), "players"), "")
	auth, err := identity.NewService(now, identity.NewMemoryStore(), nil, identity.Config{RPID:"localhost", AllowedOrigins:[]string{"http://localhost"}})
	if err != nil { t.Fatal(err) }
	return New(logger, protocol, authority, games, auth, http.NotFoundHandler())
}

func TestHealth(t *testing.T) {
	handler := testHandler(t); rr := httptest.NewRecorder(); handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/v1/healthz", nil)); if rr.Code != http.StatusOK { t.Fatalf("status=%d want %d", rr.Code, http.StatusOK) }
}

func TestGameStateRequiresAccount(t *testing.T) {
	handler := testHandler(t); rr := httptest.NewRecorder(); handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/v1/game/state", nil)); if rr.Code != http.StatusUnauthorized { t.Fatalf("state status=%d want %d", rr.Code, http.StatusUnauthorized) }
}

func TestProofOfPlayAuthorityRequiresAccount(t *testing.T) {
	handler := testHandler(t); rr := httptest.NewRecorder(); handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/v1/proof-of-play/me", nil)); if rr.Code != http.StatusUnauthorized { t.Fatalf("authority status=%d want %d", rr.Code, http.StatusUnauthorized) }
}
