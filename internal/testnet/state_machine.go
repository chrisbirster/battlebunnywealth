package testnet

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

// Transition is the deterministic input supplied to a replicated state machine.
// PreviousFinalized is nil at height 1 and otherwise contains the canonical
// finality certificate for height-1.
type Transition struct {
	NetworkID         string
	Height            uint64
	Operations        []Operation
	PreviousFinalized *FinalizedBlock
}

// StateMachine lets consensus validate a proposed state root without mutation,
// then commit exactly that transition once the block reaches finality.
type StateMachine interface {
	Root() string
	Preview(Transition) (string, error)
	Commit(Transition) (string, error)
}

type hashStateMachine struct{ root string }

func newHashStateMachine(genesisHash string) *hashStateMachine {
	return &hashStateMachine{root: hashText("bbw-pop-state/v1|genesis|" + genesisHash)}
}

func (s *hashStateMachine) Root() string { return s.root }
func (s *hashStateMachine) Preview(t Transition) (string, error) {
	return nextHashStateRoot(s.root, t.Operations), nil
}
func (s *hashStateMachine) Commit(t Transition) (string, error) {
	s.root = nextHashStateRoot(s.root, t.Operations)
	return s.root, nil
}

func nextHashStateRoot(previous string, ops []Operation) string {
	raw, _ := json.Marshal(ops)
	sum := sha256.Sum256(append([]byte(previous+"|"), raw...))
	return hex.EncodeToString(sum[:])
}
