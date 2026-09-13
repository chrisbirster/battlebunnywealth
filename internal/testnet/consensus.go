package testnet

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"
)

type Engine struct {
	mu         sync.Mutex
	genesis    Genesis
	validators []Validator
	finalized  []FinalizedBlock
	headHash   string
	stateRoot  string
	machine    StateMachine
	proposals  map[string]Proposal
	votes      map[string]map[string]Vote
	seenVotes  map[string]string
}

func NewEngine(genesis Genesis) (*Engine, error) { return NewEngineWithStateMachine(genesis, nil) }

func NewEngineWithStateMachine(genesis Genesis, machine StateMachine) (*Engine, error) {
	if err := genesis.Validate(); err != nil {
		return nil, err
	}
	if machine == nil {
		machine = newHashStateMachine(genesis.Hash)
	}
	if machine.Root() == "" {
		return nil, errors.New("state machine root required")
	}
	return &Engine{
		genesis:    genesis,
		validators: append([]Validator(nil), genesis.Validators...),
		headHash:   genesis.Hash,
		stateRoot:  machine.Root(),
		machine:    machine,
		proposals:  map[string]Proposal{},
		votes:      map[string]map[string]Vote{},
		seenVotes:  map[string]string{},
	}, nil
}

func (e *Engine) Genesis() Genesis {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.genesis
}

func (e *Engine) Validators() []Validator {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.validatorsForHeightLocked(uint64(len(e.finalized)) + 1)
}

func (e *Engine) ValidatorsAtHeight(height uint64) []Validator {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.validatorsForHeightLocked(height)
}

func (e *Engine) ProtocolVersionAtHeight(height uint64) int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.protocolVersionForHeightLocked(height)
}

func (e *Engine) Height() uint64 {
	e.mu.Lock()
	defer e.mu.Unlock()
	return uint64(len(e.finalized))
}

func (e *Engine) Finalized() []FinalizedBlock {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]FinalizedBlock(nil), e.finalized...)
}

func (e *Engine) transitionLocked(height uint64, ops []Operation) Transition {
	var previous *FinalizedBlock
	if len(e.finalized) > 0 {
		p := e.finalized[len(e.finalized)-1]
		previous = &p
	}
	return Transition{
		NetworkID:         e.genesis.NetworkID,
		Height:            height,
		Operations:        append([]Operation(nil), ops...),
		PreviousFinalized: previous,
	}
}

func (e *Engine) validatorsForHeightLocked(height uint64) []Validator {
	if rules, ok := e.machine.(ConsensusRules); ok {
		if validators := rules.ValidatorsForHeight(height); len(validators) > 0 {
			return append([]Validator(nil), validators...)
		}
	}
	return append([]Validator(nil), e.validators...)
}

func (e *Engine) protocolVersionForHeightLocked(height uint64) int {
	if rules, ok := e.machine.(ConsensusRules); ok {
		if version := rules.ProtocolVersionForHeight(height); version > 0 {
			return version
		}
	}
	return ProtocolVersion
}

func (e *Engine) validateProtocolVersionLocked(height uint64, got int) error {
	expected := e.protocolVersionForHeightLocked(height)
	if expected < ProtocolVersion || expected > MaxSupportedProtocolVersion {
		return fmt.Errorf("scheduled protocol version %d is unsupported by this binary", expected)
	}
	if got != expected {
		return fmt.Errorf("unexpected block protocol version %d at height %d; want %d", got, height, expected)
	}
	return nil
}

func (e *Engine) statusLocked() Status {
	height := uint64(len(e.finalized))
	validators := e.validatorsForHeightLocked(height + 1)
	committee := SelectCommittee(validators, e.genesis.Config.CommitteeTarget, e.headHash, height+1, 0)
	pending := ""
	for hash := range e.proposals {
		pending = hash
		break
	}
	return Status{
		NetworkID:       e.genesis.NetworkID,
		GenesisHash:     e.genesis.Hash,
		Height:          height,
		FinalizedHash:   e.headHash,
		StateRoot:       e.stateRoot,
		CommitteeSize:   len(committee),
		Quorum:          QuorumFor(e.genesis.Config, len(committee)),
		PendingProposal: pending,
	}
}

func (e *Engine) Status() Status {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.statusLocked()
}

func (e *Engine) Draft(round uint32, proposerID string, operations []Operation, now time.Time) (Block, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	height := uint64(len(e.finalized)) + 1
	version := e.protocolVersionForHeightLocked(height)
	if version < ProtocolVersion || version > MaxSupportedProtocolVersion {
		return Block{}, fmt.Errorf("scheduled protocol version %d is unsupported by this binary", version)
	}
	validators := e.validatorsForHeightLocked(height)
	committee := SelectCommittee(validators, e.genesis.Config.CommitteeTarget, e.headHash, height, round)
	proposer, ok := ExpectedProposer(committee, height, round)
	if !ok || proposer.ID != proposerID {
		return Block{}, ErrInvalidProposer
	}
	ops := append([]Operation(nil), operations...)
	if len(ops) == 0 {
		ops = []Operation{{Type: "noop"}}
	}
	root, err := e.machine.Preview(e.transitionLocked(height, ops))
	if err != nil {
		return Block{}, fmt.Errorf("preview state transition: %w", err)
	}
	block := Block{
		Version:       version,
		NetworkID:     e.genesis.NetworkID,
		Height:        height,
		Round:         round,
		PreviousHash:  e.headHash,
		PreviousState: e.stateRoot,
		StateRoot:     root,
		CommitteeHash: CommitteeHash(committee),
		ProposerID:    proposerID,
		TimestampUnix: now.UTC().Unix(),
		Operations:    ops,
	}
	block.Hash = hashBlock(block)
	return block, nil
}

func BuildProposal(block Block, validator Validator, privateKey any) (Proposal, error) {
	if block.ProposerID != validator.ID {
		return Proposal{}, ErrInvalidProposer
	}
	sig, err := Sign(validator.Algorithm, privateKey, proposalSigningMessage(block.Hash))
	if err != nil {
		return Proposal{}, err
	}
	return Proposal{Block: block, Signature: sig}, nil
}

func BuildVote(block Block, validator Validator, privateKey any) (Vote, error) {
	v := Vote{
		NetworkID:   block.NetworkID,
		Height:      block.Height,
		Round:       block.Round,
		BlockHash:   block.Hash,
		ValidatorID: validator.ID,
		Decision:    "commit",
	}
	sig, err := Sign(validator.Algorithm, privateKey, voteSigningMessage(v))
	if err != nil {
		return Vote{}, err
	}
	v.Signature = sig
	return v, nil
}

func (e *Engine) HandleProposal(p Proposal) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.handleProposalLocked(p)
}

func (e *Engine) handleProposalLocked(p Proposal) error {
	b := p.Block
	if b.NetworkID != e.genesis.NetworkID {
		return ErrWrongNetwork
	}
	expectedHeight := uint64(len(e.finalized)) + 1
	if b.Height != expectedHeight {
		return ErrInvalidHeight
	}
	if err := e.validateProtocolVersionLocked(b.Height, b.Version); err != nil {
		return err
	}
	if b.PreviousHash != e.headHash {
		return ErrInvalidPreviousHash
	}
	if b.PreviousState != e.stateRoot {
		return ErrInvalidStateRoot
	}
	root, err := e.machine.Preview(e.transitionLocked(b.Height, b.Operations))
	if err != nil {
		return fmt.Errorf("invalid state transition: %w", err)
	}
	if b.StateRoot != root {
		return ErrInvalidStateRoot
	}
	if hashBlock(b) != b.Hash {
		return errors.New("block hash mismatch")
	}
	validators := e.validatorsForHeightLocked(b.Height)
	committee := SelectCommittee(validators, e.genesis.Config.CommitteeTarget, e.headHash, b.Height, b.Round)
	if b.CommitteeHash != CommitteeHash(committee) {
		return ErrInvalidCommittee
	}
	proposer, ok := ExpectedProposer(committee, b.Height, b.Round)
	if !ok || proposer.ID != b.ProposerID {
		return ErrInvalidProposer
	}
	if !VerifySignature(proposer.Algorithm, proposer.PublicKey, proposalSigningMessage(b.Hash), p.Signature) {
		return ErrInvalidSignature
	}
	e.proposals[b.Hash] = p
	return nil
}

func (e *Engine) HandleVote(v Vote) (bool, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if v.NetworkID != e.genesis.NetworkID {
		return false, ErrWrongNetwork
	}
	expectedHeight := uint64(len(e.finalized)) + 1
	if v.Height != expectedHeight {
		return false, ErrInvalidHeight
	}
	if v.Decision != "commit" {
		return false, errors.New("invalid vote")
	}
	validators := e.validatorsForHeightLocked(v.Height)
	committee := SelectCommittee(validators, e.genesis.Config.CommitteeTarget, e.headHash, v.Height, v.Round)
	validator, ok := findValidator(committee, v.ValidatorID)
	if !ok {
		return false, ErrUnknownValidator
	}
	if !VerifySignature(validator.Algorithm, validator.PublicKey, voteSigningMessage(v), v.Signature) {
		return false, ErrInvalidSignature
	}
	seenKey := fmt.Sprintf("%d/%d/%s", v.Height, v.Round, v.ValidatorID)
	if prior, ok := e.seenVotes[seenKey]; ok {
		if prior != v.BlockHash {
			return false, ErrEquivocation
		}
		return false, ErrDuplicateVote
	}
	p, ok := e.proposals[v.BlockHash]
	if !ok {
		return false, errors.New("proposal not found")
	}
	if p.Block.Round != v.Round {
		return false, errors.New("invalid vote round")
	}
	e.seenVotes[seenKey] = v.BlockHash
	bucket := e.votes[v.BlockHash]
	if bucket == nil {
		bucket = map[string]Vote{}
		e.votes[v.BlockHash] = bucket
	}
	bucket[v.ValidatorID] = v
	quorum := QuorumFor(e.genesis.Config, len(committee))
	if len(bucket) < quorum {
		return false, nil
	}
	votes := make([]Vote, 0, len(bucket))
	for _, vote := range bucket {
		votes = append(votes, vote)
	}
	sort.Slice(votes, func(i, j int) bool { return votes[i].ValidatorID < votes[j].ValidatorID })
	cert := FinalityCertificate{
		Height:        v.Height,
		Round:         v.Round,
		BlockHash:     v.BlockHash,
		CommitteeHash: p.Block.CommitteeHash,
		Quorum:        quorum,
		Votes:         votes,
	}
	finalized := FinalizedBlock{Block: p.Block, ProposalSignature: p.Signature, Certificate: cert}
	if err := e.verifyFinalizedLocked(finalized); err != nil {
		return false, err
	}
	root, err := e.machine.Commit(e.transitionLocked(p.Block.Height, p.Block.Operations))
	if err != nil {
		return false, err
	}
	if root != p.Block.StateRoot {
		return false, ErrInvalidStateRoot
	}
	e.finalized = append(e.finalized, finalized)
	e.headHash = p.Block.Hash
	e.stateRoot = root
	e.proposals = map[string]Proposal{}
	e.votes = map[string]map[string]Vote{}
	e.seenVotes = map[string]string{}
	return true, nil
}

func (e *Engine) ImportFinalized(block FinalizedBlock) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if err := e.verifyFinalizedLocked(block); err != nil {
		return err
	}
	root, err := e.machine.Commit(e.transitionLocked(block.Block.Height, block.Block.Operations))
	if err != nil {
		return err
	}
	if root != block.Block.StateRoot {
		return ErrInvalidStateRoot
	}
	e.finalized = append(e.finalized, block)
	e.headHash = block.Block.Hash
	e.stateRoot = root
	e.proposals = map[string]Proposal{}
	e.votes = map[string]map[string]Vote{}
	e.seenVotes = map[string]string{}
	return nil
}

func (e *Engine) verifyFinalizedLocked(f FinalizedBlock) error {
	b := f.Block
	if b.NetworkID != e.genesis.NetworkID {
		return ErrWrongNetwork
	}
	if b.Height != uint64(len(e.finalized))+1 {
		return ErrInvalidHeight
	}
	if err := e.validateProtocolVersionLocked(b.Height, b.Version); err != nil {
		return err
	}
	if b.PreviousHash != e.headHash {
		return ErrInvalidPreviousHash
	}
	if b.PreviousState != e.stateRoot {
		return ErrInvalidStateRoot
	}
	root, err := e.machine.Preview(e.transitionLocked(b.Height, b.Operations))
	if err != nil {
		return fmt.Errorf("invalid state transition: %w", err)
	}
	if b.StateRoot != root {
		return ErrInvalidStateRoot
	}
	if hashBlock(b) != b.Hash {
		return errors.New("block hash mismatch")
	}
	validators := e.validatorsForHeightLocked(b.Height)
	committee := SelectCommittee(validators, e.genesis.Config.CommitteeTarget, e.headHash, b.Height, b.Round)
	if b.CommitteeHash != CommitteeHash(committee) || f.Certificate.CommitteeHash != b.CommitteeHash {
		return ErrInvalidCommittee
	}
	proposer, ok := ExpectedProposer(committee, b.Height, b.Round)
	if !ok || proposer.ID != b.ProposerID {
		return ErrInvalidProposer
	}
	if !VerifySignature(proposer.Algorithm, proposer.PublicKey, proposalSigningMessage(b.Hash), f.ProposalSignature) {
		return ErrInvalidSignature
	}
	quorum := QuorumFor(e.genesis.Config, len(committee))
	if f.Certificate.Height != b.Height || f.Certificate.Round != b.Round || f.Certificate.BlockHash != b.Hash || f.Certificate.Quorum != quorum {
		return ErrInsufficientQuorum
	}
	seen := map[string]struct{}{}
	valid := 0
	for _, vote := range f.Certificate.Votes {
		if vote.BlockHash != b.Hash || vote.Height != b.Height || vote.Round != b.Round || vote.Decision != "commit" {
			continue
		}
		if _, dup := seen[vote.ValidatorID]; dup {
			continue
		}
		validator, ok := findValidator(committee, vote.ValidatorID)
		if !ok {
			continue
		}
		if !VerifySignature(validator.Algorithm, validator.PublicKey, voteSigningMessage(vote), vote.Signature) {
			continue
		}
		seen[vote.ValidatorID] = struct{}{}
		valid++
	}
	if valid < quorum {
		return ErrInsufficientQuorum
	}
	return nil
}

func hashBlock(b Block) string {
	b.Hash = ""
	raw, _ := json.Marshal(b)
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func hashText(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func findValidator(items []Validator, id string) (Validator, bool) {
	for _, v := range items {
		if v.ID == id {
			return v, true
		}
	}
	return Validator{}, false
}
