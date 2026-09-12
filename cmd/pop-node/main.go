package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/chrisbirster/battlebunnywealth/internal/testnet"
)

func main() {
	var (
		configPath = flag.String("config", "", "path to node JSON config")
		dataDir    = flag.String("data", "data/pop-node", "node data directory")
		keyPath    = flag.String("key", "", "node Ed25519 private key path; defaults to <data>/node.key")
		keygen     = flag.String("keygen", "", "generate a node key at this path and exit")
	)
	flag.Parse()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if *keygen != "" {
		key, err := testnet.GenerateNodeKey()
		if err != nil {
			fatal(err)
		}
		if err := os.MkdirAll(filepath.Dir(*keygen), 0o700); err != nil {
			fatal(err)
		}
		if err := testnet.SaveNodeKey(*keygen, key); err != nil {
			fatal(err)
		}
		out := map[string]string{"nodeId": key.ID(), "publicKey": key.PublicText(), "keyPath": *keygen}
		_ = json.NewEncoder(os.Stdout).Encode(out)
		return
	}
	if *configPath == "" {
		fatal(errors.New("-config is required"))
	}
	cfg, err := testnet.LoadNodeConfig(*configPath)
	if err != nil {
		fatal(err)
	}
	if *keyPath == "" {
		*keyPath = filepath.Join(*dataDir, "node.key")
	}
	key, err := testnet.LoadNodeKey(*keyPath)
	if err != nil {
		fatal(fmt.Errorf("load node key: %w (run -keygen %s first)", err, *keyPath))
	}
	engine, err := testnet.NewEngine(cfg.Genesis)
	if err != nil {
		fatal(err)
	}
	store, err := testnet.OpenStore(*dataDir, cfg.Genesis)
	if err != nil {
		fatal(err)
	}
	if err := store.Restore(engine); err != nil {
		fatal(err)
	}
	guard, err := testnet.NewPeerGuard(cfg.Genesis.NetworkID, time.Duration(cfg.Genesis.Config.PeerReplayWindowSeconds)*time.Second, cfg.Peers)
	if err != nil {
		fatal(err)
	}
	node := testnet.NewHTTPNode(engine, guard, store)
	node.Relayer = &testnet.Relayer{NetworkID: cfg.Genesis.NetworkID, Key: key, Peers: cfg.Peers}
	server := &http.Server{Addr: cfg.Listen, Handler: node, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 32 << 10}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	go syncLoop(ctx, logger, cfg, engine)
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdown)
	}()
	logger.Info("Proof-of-Play node listening", "nodeId", key.ID(), "network", cfg.Genesis.NetworkID, "addr", cfg.Listen, "height", engine.Height())
	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		fatal(err)
	}
}

func syncLoop(ctx context.Context, logger *slog.Logger, cfg testnet.NodeConfig, engine *testnet.Engine) {
	ticker := time.NewTicker(cfg.SyncInterval())
	defer ticker.Stop()
	syncOnce := func() {
		for _, peer := range cfg.Peers {
			if err := testnet.SyncFrom(peer.URL, nil, engine); err != nil {
				logger.Debug("peer sync failed", "peer", peer.NodeID, "error", err)
			}
		}
	}
	syncOnce()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			syncOnce()
		}
	}
}
func fatal(err error) { fmt.Fprintln(os.Stderr, "pop-node:", err); os.Exit(1) }
