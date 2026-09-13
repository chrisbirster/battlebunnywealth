package testnet

import (
	"strings"
	"testing"
	"time"

	"github.com/chrisbirster/battlebunnywealth/internal/carrot"
)

func TestGenesisCommitsCARROTPolicy(t *testing.T) {
	pub, _, err := GenerateValidatorKey(AlgorithmEd25519)
	if err != nil {
		t.Fatal(err)
	}

	genesis, err := NewGenesis(
		"carrot-testnet",
		time.Unix(1_800_000_000, 0).UTC(),
		DefaultProtocolConfig(),
		[]Validator{{
			ID:        "validator-1",
			Algorithm: AlgorithmEd25519,
			PublicKey: pub,
			Authority: 1000,
		}},
	)
	if err != nil {
		t.Fatal(err)
	}
	if genesis.Version != ProtocolVersion {
		t.Fatalf("version=%d protocol=%d", genesis.Version, ProtocolVersion)
	}
	if ProtocolVersion != 3 {
		t.Fatalf("v0.12 requires protocol v3; got %d", ProtocolVersion)
	}
	if genesis.CarrotPolicyHash != carrot.DefaultPolicy().Hash() {
		t.Fatalf("policy hash=%s", genesis.CarrotPolicyHash)
	}

	mutated := genesis
	mutated.CarrotPolicyHash = strings.Repeat("0", 64)
	mutated.Hash = hashGenesis(mutated)
	if err := mutated.Validate(); err == nil || !strings.Contains(err.Error(), "CARROT policy") {
		t.Fatalf("expected CARROT policy mismatch, got %v", err)
	}
}
