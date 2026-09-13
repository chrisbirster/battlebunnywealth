package proofofplay

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

const PlayIntegrityOAuthScope = "https://www.googleapis.com/auth/playintegrity"

type GoogleServiceAccountCredentials struct {
	ClientEmail string `json:"client_email"`
	PrivateKey  string `json:"private_key"`
	TokenURI    string `json:"token_uri"`
}

type GoogleServiceAccountTokenSource struct {
	mu         sync.Mutex
	email      string
	privateKey *rsa.PrivateKey
	tokenURI   string
	scopes     []string
	client     *http.Client
	now        func() time.Time
	token      string
	expiry     time.Time
}

func NewGoogleServiceAccountTokenSource(raw []byte, client *http.Client, scopes ...string) (*GoogleServiceAccountTokenSource, error) {
	var credentials GoogleServiceAccountCredentials
	if err := json.Unmarshal(raw, &credentials); err != nil {
		return nil, fmt.Errorf("decode Google service account: %w", err)
	}
	if credentials.ClientEmail == "" || credentials.PrivateKey == "" {
		return nil, errors.New("Google service account client_email and private_key are required")
	}
	block, _ := pem.Decode([]byte(credentials.PrivateKey))
	if block == nil {
		return nil, errors.New("decode Google service account private key PEM")
	}
	var key *rsa.PrivateKey
	if parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		var ok bool
		key, ok = parsed.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("Google service account key is not RSA")
		}
	} else if parsed, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		key = parsed
	} else {
		return nil, errors.New("parse Google service account RSA private key")
	}
	if credentials.TokenURI == "" {
		credentials.TokenURI = "https://oauth2.googleapis.com/token"
	}
	if len(scopes) == 0 {
		scopes = []string{PlayIntegrityOAuthScope}
	}
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &GoogleServiceAccountTokenSource{email: credentials.ClientEmail, privateKey: key, tokenURI: credentials.TokenURI, scopes: append([]string(nil), scopes...), client: client, now: time.Now}, nil
}
func NewGoogleServiceAccountTokenSourceFile(path string, client *http.Client, scopes ...string) (*GoogleServiceAccountTokenSource, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return NewGoogleServiceAccountTokenSource(raw, client, scopes...)
}

func (s *GoogleServiceAccountTokenSource) AccessToken(ctx context.Context) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now().UTC()
	if s.token != "" && s.expiry.After(now.Add(time.Minute)) {
		return s.token, nil
	}
	assertion, err := s.jwtAssertion(now)
	if err != nil {
		return "", err
	}
	form := url.Values{"grant_type": {"urn:ietf:params:oauth:grant-type:jwt-bearer"}, "assertion": {assertion}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.tokenURI, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("Google OAuth token status %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	var payload struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
		TokenType   string `json:"token_type"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return "", fmt.Errorf("decode Google OAuth token: %w", err)
	}
	if payload.AccessToken == "" {
		return "", errors.New("Google OAuth response missing access_token")
	}
	if payload.ExpiresIn <= 0 {
		payload.ExpiresIn = 3600
	}
	s.token = payload.AccessToken
	s.expiry = now.Add(time.Duration(payload.ExpiresIn) * time.Second)
	return s.token, nil
}
func (s *GoogleServiceAccountTokenSource) jwtAssertion(now time.Time) (string, error) {
	header, _ := json.Marshal(map[string]string{"alg": "RS256", "typ": "JWT"})
	claims, _ := json.Marshal(map[string]any{"iss": s.email, "scope": strings.Join(s.scopes, " "), "aud": s.tokenURI, "iat": now.Unix(), "exp": now.Add(time.Hour).Unix()})
	unsigned := base64.RawURLEncoding.EncodeToString(header) + "." + base64.RawURLEncoding.EncodeToString(claims)
	digest := sha256.Sum256([]byte(unsigned))
	sig, err := rsa.SignPKCS1v15(rand.Reader, s.privateKey, crypto.SHA256, digest[:])
	if err != nil {
		return "", err
	}
	return unsigned + "." + base64.RawURLEncoding.EncodeToString(sig), nil
}
