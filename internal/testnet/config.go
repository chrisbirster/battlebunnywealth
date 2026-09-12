package testnet

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

type NodeConfig struct {
	Listen              string  `json:"listen"`
	Genesis             Genesis `json:"genesis"`
	Peers               []Peer  `json:"peers"`
	SyncIntervalSeconds int     `json:"syncIntervalSeconds"`
}

func LoadNodeConfig(path string) (NodeConfig, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return NodeConfig{}, err
	}
	var cfg NodeConfig
	decErr := json.Unmarshal(raw, &cfg)
	if decErr != nil {
		return NodeConfig{}, fmt.Errorf("decode node config: %w", decErr)
	}
	if err := cfg.Genesis.Validate(); err != nil {
		return NodeConfig{}, fmt.Errorf("invalid genesis: %w", err)
	}
	if strings.TrimSpace(cfg.Listen) == "" {
		cfg.Listen = ":9101"
	}
	if cfg.SyncIntervalSeconds <= 0 {
		cfg.SyncIntervalSeconds = 5
	}
	seen := map[string]struct{}{}
	for _, peer := range cfg.Peers {
		if peer.NodeID == "" || peer.PublicKey == "" || peer.URL == "" {
			return NodeConfig{}, errors.New("peer nodeId, publicKey, and url are required")
		}
		if _, ok := seen[peer.NodeID]; ok {
			return NodeConfig{}, fmt.Errorf("duplicate peer %q", peer.NodeID)
		}
		seen[peer.NodeID] = struct{}{}
	}
	return cfg, nil
}

func (c NodeConfig) SyncInterval() time.Duration {
	return time.Duration(c.SyncIntervalSeconds) * time.Second
}
