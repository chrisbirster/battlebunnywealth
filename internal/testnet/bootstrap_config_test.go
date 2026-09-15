package testnet

import (
	"fmt"
	"testing"
	"time"
)

func TestBuildNodeConfigs(t *testing.T) {
	validators := make([]Validator, 4)
	for i := range validators {
		pub, _, err := GenerateValidatorKey(AlgorithmEd25519)
		if err != nil { t.Fatal(err) }
		validators[i] = Validator{ID: fmt.Sprintf("validator-%d", i+1), PublicKey: pub, Algorithm: AlgorithmEd25519, Authority: 1000}
	}
	genesis, err := NewGenesis("bootstrap-config-test", time.Unix(1_800_000_000, 0).UTC(), DefaultProtocolConfig(), validators)
	if err != nil { t.Fatal(err) }

	names := []string{"iad", "ord", "dfw", "lax"}
	nodes := make([]BootstrapNode, 0, len(names))
	for _, name := range names {
		key, err := GenerateNodeKey()
		if err != nil { t.Fatal(err) }
		nodes = append(nodes, BootstrapNode{Name: name, NodeID: key.ID(), PublicKey: key.PublicText(), URL: "https://bbwealth-pop-" + name + ".fly.dev"})
	}
	configs, err := BuildNodeConfigs(genesis, nodes, 5)
	if err != nil { t.Fatal(err) }
	if len(configs) != 4 { t.Fatalf("configs=%d", len(configs)) }
	for _, name := range names {
		cfg := configs[name]
		if cfg.Genesis.Hash != genesis.Hash { t.Fatalf("%s genesis mismatch", name) }
		if cfg.Listen != ":9101" || cfg.SyncIntervalSeconds != 5 { t.Fatalf("%s runtime defaults wrong", name) }
		if len(cfg.Peers) != 3 { t.Fatalf("%s peers=%d", name, len(cfg.Peers)) }
	}
}

func TestBuildNodeConfigsRejectsMismatchedNodeID(t *testing.T) {
	pub, _, err := GenerateValidatorKey(AlgorithmEd25519)
	if err != nil { t.Fatal(err) }
	genesis, err := NewGenesis("bootstrap-config-test", time.Unix(1_800_000_000, 0).UTC(), DefaultProtocolConfig(), []Validator{{ID: "validator-1", PublicKey: pub, Algorithm: AlgorithmEd25519, Authority: 1000}})
	if err != nil { t.Fatal(err) }
	keyA, _ := GenerateNodeKey()
	keyB, _ := GenerateNodeKey()
	_, err = BuildNodeConfigs(genesis, []BootstrapNode{
		{Name: "iad", NodeID: "deadbeef", PublicKey: keyA.PublicText(), URL: "https://a.fly.dev"},
		{Name: "ord", NodeID: keyB.ID(), PublicKey: keyB.PublicText(), URL: "https://b.fly.dev"},
	}, 5)
	if err == nil { t.Fatal("expected node id/public key mismatch") }
}
