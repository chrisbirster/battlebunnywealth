package publictestnet

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"sort"

	"github.com/chrisbirster/battlebunnywealth/internal/carrot"
	"github.com/chrisbirster/battlebunnywealth/internal/testnet"
)

const CheckpointVersion = 1

var ErrInvalidCheckpoint = errors.New("invalid public-testnet checkpoint")

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

// VerifyCheckpoint independently reconstructs consensus from genesis and the
// finalized block sequence. The checkpoint is therefore a compact comparison
// target, not a trusted snapshot that bypasses block/finality verification.
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
