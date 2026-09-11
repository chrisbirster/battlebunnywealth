package httpserver

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/chrisbirster/battlebunnywealth/internal/proofofplay"
)

func TestHealth(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	protocol := proofofplay.NewProtocol(proofofplay.DefaultConfig(), proofofplay.NewChain(time.Unix(0, 0)))
	handler := New(logger, protocol, http.NotFoundHandler())

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/v1/healthz", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d want %d", rr.Code, http.StatusOK)
	}
}
