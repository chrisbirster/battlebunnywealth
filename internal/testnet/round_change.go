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
}

type RoundProgress struct {
	Height       uint64             `json:"height"`
	Round        uint32             `json:"round"`
	Locks        []ValidatorLock    `json:"locks,omitempty"`
	Certificates []RoundCertificate `json:"certificates,omitempty"`
}

func BuildRoundChange(networkID string, height uint64, fromRound uint32, validator Validator, privateKey any, lockedRound uint32, lockedValueHash string) (RoundChange, error) {
	if networkID == "" || height == 0 || validator.ID == "" {
		return RoundChange{}, ErrInvalidRoundChange
	}
	if lockedValueHash == "" {
		lockedRound = 0
	}
	rc := RoundChange{NetworkID: networkID, Height: height, FromRound: fromRound, ToRound: fromRound + 1, ValidatorID: validator.ID, LockedRound: lockedRound, LockedValueHash: lockedValueHash}
	sig, err := Sign(validator.Algorithm, privateKey, roundChangeSigningMessage(rc))
	if err != nil {
		return RoundChange{}, err
	}
	rc.Signature = sig
	return rc, nil
}

func roundChangeSigningMessage(rc RoundChange) string {
	return fmt.Sprintf("bbw-round-change/v1\nnetwork=%s\nheight=%d\nfromRound=%d\ntoRound=%d\nvalidator=%s\nlockedRound=%d\nlockedValue=%s", rc.NetworkID, rc.Height, rc.FromRound, rc.ToRound, rc.ValidatorID, rc.LockedRound, rc.LockedValueHash)
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

func chooseCertificateLock(changes []RoundChange) (uint32, string, error) {
	var highest uint32
	value := ""
	for _, change := range changes {
		if change.LockedValueHash == "" {
			continue
		}
		if change.LockedRound > highest {
			highest = change.LockedRound
			value = change.LockedValueHash
			continue
		}
		if change.LockedRound == highest && value != "" && change.LockedValueHash != value {
			return 0, "", ErrConflictingLocks
		}
		if change.LockedRound == highest && value == "" {
			value = change.LockedValueHash
		}
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
