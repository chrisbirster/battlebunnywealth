package proofofplay

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestGoogleServiceAccountTokenSourceRefreshesAndCaches(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	privatePEM := string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}))
	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		if err := r.ParseForm(); err != nil {
			t.Errorf("parse form: %v", err)
			http.Error(w, "bad", 400)
			return
		}
		if got := r.Form.Get("grant_type"); got != "urn:ietf:params:oauth:grant-type:jwt-bearer" {
			t.Errorf("grant=%q", got)
		}
		if parts := strings.Split(r.Form.Get("assertion"), "."); len(parts) != 3 {
			t.Errorf("jwt parts=%d", len(parts))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"token-1","expires_in":3600,"token_type":"Bearer"}`))
	}))
	defer server.Close()
	credentials, _ := json.Marshal(GoogleServiceAccountCredentials{ClientEmail: "service@example.iam.gserviceaccount.com", PrivateKey: privatePEM, TokenURI: server.URL})
	source, err := NewGoogleServiceAccountTokenSource(credentials, server.Client())
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC)
	source.now = func() time.Time { return now }
	first, err := source.AccessToken(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	second, err := source.AccessToken(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if first != "token-1" || second != first {
		t.Fatalf("tokens %q %q", first, second)
	}
	if hits.Load() != 1 {
		t.Fatalf("token endpoint hits=%d", hits.Load())
	}
}
