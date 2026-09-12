package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/chrisbirster/battlebunnywealth/internal/game"
	"github.com/chrisbirster/battlebunnywealth/internal/httpserver"
	"github.com/chrisbirster/battlebunnywealth/internal/identity"
	"github.com/chrisbirster/battlebunnywealth/internal/proofofplay"
	webapp "github.com/chrisbirster/battlebunnywealth/web"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	addr := envOr("BBWEALTH_ADDR", ":8080")
	origins := splitCSV(envOr("BBWEALTH_ORIGINS", "http://localhost:8080,http://localhost:5173"))

	chain := proofofplay.NewChain(time.Unix(0, 0).UTC())
	protocol := proofofplay.NewProtocol(proofofplay.DefaultConfig(), chain)
	authority, err := proofofplay.NewAuthorityService(time.Now, proofofplay.NewAuthorityFileStore(envOr("BBWEALTH_POP_STATE", "data/proof-of-play.json")), proofofplay.DefaultAuthorityConfig())
	if err != nil { logger.Error("open Proof-of-Play authority state", "error", err); os.Exit(1) }

	attConfig := proofofplay.DefaultAttestationConfig()
	attConfig.AppleBundleID = os.Getenv("BBWEALTH_APPLE_BUNDLE_ID")
	attConfig.AppleTeamID = os.Getenv("BBWEALTH_APPLE_TEAM_ID")
	attConfig.AppleEnvironment = envOr("BBWEALTH_APPLE_ATTEST_ENV", "development")
	attConfig.AndroidPackageName = os.Getenv("BBWEALTH_ANDROID_PACKAGE")
	attConfig.AndroidAllowedCertificates = splitCSV(os.Getenv("BBWEALTH_ANDROID_CERTIFICATES"))
	attConfig.AndroidRequireStrongIntegrity = os.Getenv("BBWEALTH_ANDROID_REQUIRE_STRONG_INTEGRITY") == "1"
	providers := []proofofplay.AttestationProviderVerifier{
		proofofplay.AppleAppAttestVerifier{Config: attConfig},
		proofofplay.GooglePlayIntegrityVerifier{Config: attConfig, Decoder: proofofplay.PlayIntegrityHTTPDecoder{Tokens: proofofplay.StaticAccessToken(os.Getenv("BBWEALTH_PLAY_INTEGRITY_ACCESS_TOKEN"))}},
	}
	if os.Getenv("BBWEALTH_DEV_CONTROLS") == "1" { providers = append(providers, proofofplay.DevelopmentAttestationVerifier{}) }
	attestation, err := proofofplay.NewAttestationService(time.Now, proofofplay.NewAttestationFileStore(envOr("BBWEALTH_ATTESTATION_STATE", "data/attestation.json")), attConfig, providers...)
	if err != nil { logger.Error("open Proof-of-Play attestation state", "error", err); os.Exit(1) }

	gameRegistry := game.NewRegistry(time.Now, envOr("BBWEALTH_GAME_DIR", "data/players"), envOr("BBWEALTH_GAME_STATE", "data/game-state.json"))
	identityService, err := identity.NewService(time.Now, identity.NewFileStore(envOr("BBWEALTH_IDENTITY_STATE", "data/identity.json")), identity.NewPLCResolver(nil), identity.Config{
		RPID: envOr("BBWEALTH_RP_ID", "localhost"), RPName: "Battle Bunny Wealth", AllowedOrigins: origins, SessionTTL: 30 * 24 * time.Hour,
	})
	if err != nil { logger.Error("open identity state", "error", err); os.Exit(1) }
	spa, err := webapp.Handler()
	if err != nil { logger.Error("create spa handler", "error", err); os.Exit(1) }

	base := httpserver.New(logger, protocol, authority, gameRegistry, identityService, spa)
	handler := httpserver.WithAttestation(base, protocol, authority, attestation, identityService)
	server := &http.Server{Addr: addr, Handler: handler, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	go func() { <-ctx.Done(); shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second); defer cancel(); if err := server.Shutdown(shutdownCtx); err != nil { logger.Error("shutdown server", "error", err) } }()

	logger.Info("battle bunny wealth listening", "addr", addr)
	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) { logger.Error("server stopped", "error", err); os.Exit(1) }
}

func envOr(key, fallback string) string { if value := os.Getenv(key); value != "" { return value }; return fallback }
func splitCSV(value string) []string { var out []string; for _, item := range strings.Split(value, ",") { if item = strings.TrimSpace(item); item != "" { out = append(out, item) } }; return out }
