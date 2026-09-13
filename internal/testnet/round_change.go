package testnet

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
)

var (
	ErrStaleRound         = errors.New("stale consensus round")
	ErrFutureRound        = errors.New("future consensus round has no certificate")
	ErrInvalidRoundChange = errors.New("invalid round-change message")
	ErrConflictingLocks   = errors.New("conflicting validator locks")
	ErrLockedValue        = errors.New("validator is locked on another value")
)

type RoundChange struct {
	NetworkID       string `json:"networkId"`
	Height          uint64 `json:"height"`
	FromRound       uint32 `json:"fromRound"`
	ToRound         uint32 `json:"toRound"`
	ValidatorID     string `json:"validatorId"`
	LockedRound     uint32 `json:"lockedRound,omitempty"`
	LockedValueHash string `json:"lockedValueHash,omitempty"`
	LockProof       *Vote  `json:"lockProof,omitempty"`
	Signature       string `json:"signature"`
}

type RoundCertificate struct {
	NetworkID       string        `json:"networkId"`
	Height          uint64        `json:"height"`
	Round           uint32        `json:"round"`
	CommitteeHash   string        `json:"committeeHash"`
	Quorum          int           `json:"quorum"`
	LockedRound     uint32        `json:"lockedRound,omitempty"`
	LockedValueHash string        `json:"lockedValueHash,omitempty"`
	Changes         []RoundChange `json:"changes"`
}

type ValidatorLock struct {
	ValidatorID string `json:"validatorId"`
	Round       uint32 `json:"round"`
	ValueHash   string `json:"valueHash"`
	Proof       Vote   `json:"proof"`
}

type RoundProgress struct {
	Height       uint64             `json:"height"`
	Round        uint32             `json:"round"`
	Locks        []ValidatorLock    `json:"locks,omitempty"`
	Certificates []RoundCertificate `json:"certificates,omitempty"`
}

// BuildRoundChange keeps an optional variadic proof so unlocked callers remain
// source-compatible. A lock claim, however, requires exactly one signed vote
// proving that this validator actually voted for the claimed consensus value.
func BuildRoundChange(networkID string, height uint64, fromRound uint32, validator Validator, privateKey any, lockedRound uint32, lockedValueHash string, proof ...Vote) (RoundChange, error) {
	if networkID == "" || height == 0 || validator.ID == "" {
		return RoundChange{}, ErrInvalidRoundChange
	}
	rc := RoundChange{
		NetworkID:       networkID,
		Height:          height,
		FromRound:       fromRound,
		ToRound:         fromRound + 1,
		ValidatorID:     validator.ID,
		LockedRound:     lockedRound,
		LockedValueHash: lockedValueHash,
	}
	if lockedValueHash == "" {
		rc.LockedRound = 0
		if len(proof) != 0 {
			return RoundChange{}, ErrInvalidRoundChange
		}
	} else {
		if len(proof) != 1 {
			return RoundChange{}, ErrInvalidRoundChange
		}
		p := proof[0]
		if err := verifyLockProof(rc, validator, p); err != nil {
			return RoundChange{}, err
		}
		rc.LockProof = &p
	}
	sig, err := Sign(validator.Algorithm, privateKey, roundChangeSigningMessage(rc))
	if err != nil {
		return RoundChange{}, err
	}
	rc.Signature = sig
	return rc, nil
}

func roundChangeSigningMessage(rc RoundChange) string {
	proofBlock := ""
	if rc.LockProof != nil {
		proofBlock = rc.LockProof.BlockHash
	}
	return fmt.Sprintf("bbw-round-change/v2\nnetwork=%s\nheight=%d\nfromRound=%d\ntoRound=%d\nvalidator=%s\nlockedRound=%d\nlockedValue=%s\nlockProofBlock=%s", rc.NetworkID, rc.Height, rc.FromRound, rc.ToRound, rc.ValidatorID, rc.LockedRound, rc.LockedValueHash, proofBlock)
}

func verifyLockProof(change RoundChange, validator Validator, proof Vote) error {
	if change.LockedValueHash == "" {
		if change.LockProof != nil || change.LockedRound != 0 {
			return ErrInvalidRoundChange
		}
		return nil
	}
	if proof.NetworkID != change.NetworkID || proof.Height != change.Height || proof.Round != change.LockedRound || proof.ValidatorID != change.ValidatorID || proof.Decision != "commit" || proof.ValueHash == "" || proof.ValueHash != change.LockedValueHash || proof.BlockHash == "" {
		return ErrInvalidRoundChange
	}
	if !VerifySignature(validator.Algorithm, validator.PublicKey, voteSigningMessage(proof), proof.Signature) {
		return ErrInvalidSignature
	}
	return nil
}

func BlockValueHash(b Block) string { return blockValueHash(b) }

func blockValueHash(b Block) string {
	value := struct {
		Version       int         `json:"version"`
		NetworkID     string      `json:"networkId"`
		Height        uint64      `json:"height"`
		PreviousHash  string      `json:"previousHash"`
		PreviousState string      `json:"previousState"`
		StateRoot     string      `json:"stateRoot"`
		Operations    []Operation `json:"operations"`
	}{Version: b.Version, NetworkID: b.NetworkID, Height: b.Height, PreviousHash: b.PreviousHash, PreviousState: b.PreviousState, StateRoot: b.StateRoot, Operations: b.Operations}
	raw, _ := json.Marshal(value)
	return hashText(string(raw))
}

// quorumIntersectionThreshold is the minimum number of matching lock proofs
// guaranteed to occur in any quorum-sized round-change certificate if the same
// value previously obtained a finality quorum: |Q1 ∩ Q2| >= 2Q-N.
func quorumIntersectionThreshold(committeeSize, quorum int) int {
	threshold := 2*quorum - committeeSize
	if threshold < 1 {
		threshold = 1
	}
	return threshold
}

// chooseCertificateLock ignores isolated lock claims. A value is carried into
// the next round only when matching signed lock proofs reach the quorum-
// intersection threshold. This prevents one malicious validator from inventing
// a lone high-round lock that stalls the whole committee while preserving a
// value that could already have finalized on another honest peer.
func chooseCertificateLock(changes []RoundChange, minimumProofs int) (uint32, string, error) {
	if minimumProofs < 1 {
		minimumProofs = 1
	}
	type key struct {
		round uint32
		value string
	}
	counts := map[key]int{}
	for _, change := range changes {
		if change.LockedValueHash == "" {
			continue
		}
		counts[key{round: change.LockedRound, value: change.LockedValueHash}]++
	}

	var highest uint32
	value := ""
	found := false
	for candidate, count := range counts {
		if count < minimumProofs {
			continue
		}
		if !found || candidate.round > highest {
			highest = candidate.round
			value = candidate.value
			found = true
			continue
		}
		if candidate.round == highest && candidate.value != value {
			return 0, "", ErrConflictingLocks
		}
	}
	if !found {
		return 0, "", nil
	}
	return highest, value, nil
}

func sortedRoundChanges(items map[string]RoundChange) []RoundChange {
	out := make([]RoundChange, 0, len(items))
	for _, item := range items {
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ValidatorID < out[j].ValidatorID })
	return out
}
