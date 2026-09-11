package proofofplay

import (
	"errors"
	"testing"
	"time"
)

func TestChainAppend(t *testing.T) {
	at := time.Date(2026, 9, 11, 0, 0, 0, 0, time.UTC)
	chain := NewChain(at)
	head := chain.Head()
	block := NewBlock(head, 1, "bunny-1", nil, nil, at.Add(time.Minute))

	if err := chain.Append(block); err != nil {
		t.Fatalf("append block: %v", err)
	}
	if got := chain.Height(); got != 1 {
		t.Fatalf("height=%d want 1", got)
	}
}

func TestChainRejectsTamperedBlock(t *testing.T) {
	at := time.Date(2026, 9, 11, 0, 0, 0, 0, time.UTC)
	chain := NewChain(at)
	block := NewBlock(chain.Head(), 1, "bunny-1", nil, nil, at.Add(time.Minute))
	block.Header.Proposer = "evil-bunny"

	if err := chain.Append(block); !errors.Is(err, ErrInvalidBlockHash) {
		t.Fatalf("error=%v want %v", err, ErrInvalidBlockHash)
	}
}

func TestValidateEnvelope(t *testing.T) {
	now := time.Date(2026, 9, 11, 0, 0, 0, 0, time.UTC)
	participant := Participant{ID: "p1", DevicePublicKey: "key1"}
	challenge := Challenge{ID: "c1", Epoch: 9, ExpiresAt: now.Add(time.Minute)}
	proof := ParticipationProof{ParticipantID: "p1", DevicePublicKey: "key1", ChallengeID: "c1", Epoch: 9}

	if err := ValidateEnvelope(now, participant, challenge, proof); err != nil {
		t.Fatalf("validate envelope: %v", err)
	}
}
