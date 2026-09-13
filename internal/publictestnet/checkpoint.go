package publictestnet

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/chrisbirster/battlebunnywealth/internal/carrot"
	"github.com/chrisbirster/battlebunnywealth/internal/testnet"
)

const CheckpointVersion = 1

var (
	ErrInvalidCheckpoint = errors.New("invalid public-testnet checkpoint")
	ErrCheckpointMismatch = errors.New("public-testnet checkpoint mismatch")
)

type Checkpoint struct {
	Version          int    `json:"version"`
	NetworkID        string `json:"networkId"`
	GenesisHash      string `json:"genesisHash"`
	Height           uint64 `json:"height"`
	FinalizedHash    string `json:"finalizedHash"`
	StateRoot        string `json:"stateRoot"`
	ProtocolVersion  int    `json:"protocolVersion"`
	ValidatorSetHash string `json:"validatorSetHash"`
	CarrotPolicyHash string `json:"carrotPolicyHash"`
	Hash             string `json:"hash"`
}

type CheckpointObservation struct {
	Source     string     `json:"source"`
	Checkpoint Checkpoint `json:"checkpoint"`
}

type CheckpointComparison struct {
	NetworkID     string   `json:"networkId"`
	GenesisHash   string   `json:"genesisHash"`
	Height        uint64   `json:"height"`
	FinalizedHash string   `json:"finalizedHash"`
	StateRoot     string   `json:"stateRoot"`
	Sources       []string `json:"sources"`
	Match         bool     `json:"match"`
}

func BuildCheckpoint(engine *testnet.Engine, state *ConsensusState) (Checkpoint, error) {
	if engine == nil || state == nil {
		return Checkpoint{}, ErrInvalidCheckpoint
	}
	status := engine.Status()
	heightForRules := status.Height
	if heightForRules == 0 {
		heightForRules = 1
	}
	checkpoint := Checkpoint{
		Version:          CheckpointVersion,
		NetworkID:        status.NetworkID,
		GenesisHash:      status.GenesisHash,
		Height:           status.Height,
		FinalizedHash:    status.FinalizedHash,
		StateRoot:        status.StateRoot,
		ProtocolVersion:  engine.ProtocolVersionAtHeight(heightForRules),
		ValidatorSetHash: hashValidatorSet(engine.ValidatorsAtHeight(heightForRules)),
		CarrotPolicyHash: carrot.DefaultPolicy().Hash(),
	}
	checkpoint.Hash = checkpointHash(checkpoint)
	return checkpoint, nil
}

func VerifyCheckpoint(genesis testnet.Genesis, finalized []testnet.FinalizedBlock, checkpoint Checkpoint) error {
	if checkpoint.Version != CheckpointVersion || checkpoint.NetworkID != genesis.NetworkID || checkpoint.GenesisHash != genesis.Hash || checkpoint.CarrotPolicyHash != carrot.DefaultPolicy().Hash() || checkpoint.Hash != checkpointHash(checkpoint) {
		return ErrInvalidCheckpoint
	}
	if checkpoint.Height > uint64(len(finalized)) {
		return ErrInvalidCheckpoint
	}
	state, err := NewConsensusState(genesis, carrot.DefaultPolicy())
	if err != nil {
		return err
	}
	engine, err := testnet.NewEngineWithStateMachine(genesis, state)
	if err != nil {
		return err
	}
	for i := uint64(0); i < checkpoint.Height; i++ {
		if err := engine.ImportFinalized(finalized[i]); err != nil {
			return err
		}
	}
	status := engine.Status()
	heightForRules := checkpoint.Height
	if heightForRules == 0 {
		heightForRules = 1
	}
	if status.Height != checkpoint.Height || status.FinalizedHash != checkpoint.FinalizedHash || status.StateRoot != checkpoint.StateRoot || engine.ProtocolVersionAtHeight(heightForRules) != checkpoint.ProtocolVersion || hashValidatorSet(engine.ValidatorsAtHeight(heightForRules)) != checkpoint.ValidatorSetHash {
		return ErrInvalidCheckpoint
	}
	return nil
}

// CompareCheckpointSet is intentionally strict: all operator observations must
// describe the exact same height and consensus commitment. A caller should poll
// until operators reach a common height rather than silently comparing unequal
// heights and treating lag as agreement.
func CompareCheckpointSet(observations []CheckpointObservation) (CheckpointComparison, error) {
	if len(observations) < 2 {
		return CheckpointComparison{}, errors.New("at least two checkpoint observations required")
	}
	base := observations[0].Checkpoint
	if base.Version != CheckpointVersion || base.Hash != checkpointHash(base) {
		return CheckpointComparison{}, ErrInvalidCheckpoint
	}
	comparison := CheckpointComparison{NetworkID: base.NetworkID, GenesisHash: base.GenesisHash, Height: base.Height, FinalizedHash: base.FinalizedHash, StateRoot: base.StateRoot, Match: true}
	for _, observation := range observations {
		checkpoint := observation.Checkpoint
		if observation.Source == "" || checkpoint.Version != CheckpointVersion || checkpoint.Hash != checkpointHash(checkpoint) {
			return CheckpointComparison{}, ErrInvalidCheckpoint
		}
		comparison.Sources = append(comparison.Sources, observation.Source)
		if checkpoint.NetworkID != base.NetworkID || checkpoint.GenesisHash != base.GenesisHash || checkpoint.Height != base.Height || checkpoint.FinalizedHash != base.FinalizedHash || checkpoint.StateRoot != base.StateRoot || checkpoint.ProtocolVersion != base.ProtocolVersion || checkpoint.ValidatorSetHash != base.ValidatorSetHash || checkpoint.CarrotPolicyHash != base.CarrotPolicyHash {
			comparison.Match = false
		}
	}
	sort.Strings(comparison.Sources)
	if !comparison.Match {
		return comparison, fmt.Errorf("%w at height %d", ErrCheckpointMismatch, base.Height)
	}
	return comparison, nil
}

func hashValidatorSet(validators []testnet.Validator) string {
	normalized := append([]testnet.Validator(nil), validators...)
	sort.Slice(normalized, func(i, j int) bool { return normalized[i].ID < normalized[j].ID })
	raw, _ := json.Marshal(normalized)
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func checkpointHash(checkpoint Checkpoint) string {
	checkpoint.Hash = ""
	raw, _ := json.Marshal(checkpoint)
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}
