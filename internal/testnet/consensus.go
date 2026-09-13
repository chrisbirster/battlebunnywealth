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
	mu                sync.Mutex
	genesis           Genesis
	validators        []Validator
	finalized         []FinalizedBlock
	headHash          string
	stateRoot         string
	machine           StateMachine
	proposals         map[string]Proposal
	votes             map[string]map[string]Vote
	seenVotes         map[string]string
	currentRound      uint32
	roundChanges      map[uint32]map[string]RoundChange
	roundCertificates map[uint32]RoundCertificate
	validatorLocks    map[string]ValidatorLock
	proposalSlots     map[string]string
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
		genesis:           genesis,
		validators:        append([]Validator(nil), genesis.Validators...),
		headHash:          genesis.Hash,
		stateRoot:         machine.Root(),
		machine:           machine,
		proposals:         map[string]Proposal{},
		votes:             map[string]map[string]Vote{},
		seenVotes:         map[string]string{},
		roundChanges:      map[uint32]map[string]RoundChange{},
		roundCertificates: map[uint32]RoundCertificate{},
		validatorLocks:    map[string]ValidatorLock{},
		proposalSlots:     map[string]string{},
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
func (e *Engine) CurrentRound() uint32 {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.currentRound
}
func (e *Engine) RoundCertificate(round uint32) (RoundCertificate, bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	cert, ok := e.roundCertificates[round]
	return cert, ok
}
func (e *Engine) RoundProgress() RoundProgress {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.roundProgressLocked()
}

func (e *Engine) RestoreRoundProgress(progress RoundProgress) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	wantHeight := uint64(len(e.finalized)) + 1
	if progress.Height != wantHeight {
		return fmt.Errorf("round progress height %d does not match next height %d", progress.Height, wantHeight)
	}
	committee := e.committeeForHeightLocked(wantHeight)
	locks := map[string]ValidatorLock{}
	for _, lock := range progress.Locks {
		if err := e.verifyValidatorLockLocked(lock, wantHeight, committee); err != nil {
			return err
		}
		if prior, ok := locks[lock.ValidatorID]; ok && prior.ValueHash != lock.ValueHash {
			return ErrConflictingLocks
		}
		locks[lock.ValidatorID] = lock
	}
	e.validatorLocks = locks
	e.currentRound = 0
	e.roundCertificates = map[uint32]RoundCertificate{}
	certs := append([]RoundCertificate(nil), progress.Certificates...)
	sort.Slice(certs, func(i, j int) bool { return certs[i].Round < certs[j].Round })
	for _, cert := range certs {
		if cert.Round != e.currentRound+1 {
			return ErrFutureRound
		}
		if err := e.verifyRoundCertificateLocked(cert); err != nil {
			return err
		}
		e.roundCertificates[cert.Round] = cert
		e.recordCertificateLocksLocked(cert)
		e.currentRound = cert.Round
	}
	if progress.Round != e.currentRound {
		return fmt.Errorf("round progress round %d does not match certificate chain %d", progress.Round, e.currentRound)
	}
	return nil
}

func (e *Engine) transitionLocked(height uint64, ops []Operation) Transition {
	var previous *FinalizedBlock
	if len(e.finalized) > 0 {
		p := e.finalized[len(e.finalized)-1]
		previous = &p
	}
	return Transition{NetworkID: e.genesis.NetworkID, Height: height, Operations: append([]Operation(nil), ops...), PreviousFinalized: previous}
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

func (e *Engine) committeeForHeightLocked(height uint64) []Validator {
	validators := e.validatorsForHeightLocked(height)
	// Committee membership is fixed for the height. Rounds rotate only the
	// proposer, preserving quorum intersection across timeout certificates.
	return SelectCommittee(validators, e.genesis.Config.CommitteeTarget, e.headHash, height, 0)
}

func (e *Engine) statusLocked() Status {
	height := uint64(len(e.finalized))
	committee := e.committeeForHeightLocked(height + 1)
	pending := ""
	for hash := range e.proposals {
		pending = hash
		break
	}
	return Status{NetworkID: e.genesis.NetworkID, GenesisHash: e.genesis.Hash, Height: height, CurrentRound: e.currentRound, FinalizedHash: e.headHash, StateRoot: e.stateRoot, CommitteeSize: len(committee), Quorum: QuorumFor(e.genesis.Config, len(committee)), LockedValidators: len(e.validatorLocks), PendingProposal: pending}
}

func (e *Engine) Status() Status {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.statusLocked()
}

func (e *Engine) Draft(round uint32, proposerID string, operations []Operation, now time.Time) (Block, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if round < e.currentRound {
		return Block{}, ErrStaleRound
	}
	if round > e.currentRound {
		return Block{}, ErrFutureRound
	}
	height := uint64(len(e.finalized)) + 1
	version := e.protocolVersionForHeightLocked(height)
	if version < ProtocolVersion || version > MaxSupportedProtocolVersion {
		return Block{}, fmt.Errorf("scheduled protocol version %d is unsupported by this binary", version)
	}
	committee := e.committeeForHeightLocked(height)
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
	block := Block{Version: version, NetworkID: e.genesis.NetworkID, Height: height, Round: round, PreviousHash: e.headHash, PreviousState: e.stateRoot, StateRoot: root, CommitteeHash: CommitteeHash(committee), ProposerID: proposerID, TimestampUnix: now.UTC().Unix(), Operations: ops}
	if round > 0 {
		cert, ok := e.roundCertificates[round]
		if !ok {
			return Block{}, ErrFutureRound
		}
		copy := cert
		block.RoundCertificate = &copy
		if cert.LockedValueHash != "" && blockValueHash(block) != cert.LockedValueHash {
			return Block{}, ErrLockedValue
		}
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
	v := Vote{NetworkID: block.NetworkID, Height: block.Height, Round: block.Round, BlockHash: block.Hash, ValueHash: blockValueHash(block), ValidatorID: validator.ID, Decision: "commit"}
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
	if b.Round < e.currentRound {
		return ErrStaleRound
	}
	if b.Round > e.currentRound {
		if b.Round != e.currentRound+1 || b.RoundCertificate == nil {
			return ErrFutureRound
		}
		if err := e.applyRoundCertificateLocked(*b.RoundCertificate); err != nil {
			return err
		}
	}
	if b.Round > 0 {
		cert, ok := e.roundCertificates[b.Round]
		if !ok {
			return ErrFutureRound
		}
		if b.RoundCertificate != nil {
			if err := e.verifyRoundCertificateLocked(*b.RoundCertificate); err != nil {
				return err
			}
		}
		if cert.LockedValueHash != "" && blockValueHash(b) != cert.LockedValueHash {
			return ErrLockedValue
		}
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
	committee := e.committeeForHeightLocked(b.Height)
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
	slot := fmt.Sprintf("%d/%d/%s", b.Height, b.Round, b.ProposerID)
	if prior, ok := e.proposalSlots[slot]; ok && prior != b.Hash {
		return ErrEquivocation
	}
	e.proposalSlots[slot] = b.Hash
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
	if v.Round < e.currentRound {
		return false, ErrStaleRound
	}
	if v.Round > e.currentRound {
		return false, ErrFutureRound
	}
	if v.Decision != "commit" {
		return false, errors.New("invalid vote")
	}
	committee := e.committeeForHeightLocked(v.Height)
	validator, ok := findValidator(committee, v.ValidatorID)
	if !ok {
		return false, ErrUnknownValidator
	}
	if v.ValueHash == "" || !VerifySignature(validator.Algorithm, validator.PublicKey, voteSigningMessage(v), v.Signature) {
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
	valueHash := blockValueHash(p.Block)
	if v.ValueHash != valueHash {
		return false, ErrInvalidSignature
	}
	if lock, ok := e.validatorLocks[v.ValidatorID]; ok && lock.ValueHash != valueHash {
		return false, ErrLockedValue
	}
	e.validatorLocks[v.ValidatorID] = ValidatorLock{ValidatorID: v.ValidatorID, Round: v.Round, ValueHash: valueHash, Proof: v}
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
	cert := FinalityCertificate{Height: v.Height, Round: v.Round, BlockHash: v.BlockHash, CommitteeHash: p.Block.CommitteeHash, Quorum: quorum, Votes: votes}
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
	e.resetRoundLocked()
	return true, nil
}

func (e *Engine) HandleRoundChange(change RoundChange) (bool, RoundCertificate, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if change.NetworkID != e.genesis.NetworkID {
		return false, RoundCertificate{}, ErrWrongNetwork
	}
	height := uint64(len(e.finalized)) + 1
	if change.Height != height {
		return false, RoundCertificate{}, ErrInvalidHeight
	}
	if change.FromRound < e.currentRound {
		return false, RoundCertificate{}, ErrStaleRound
	}
	if change.FromRound > e.currentRound || change.ToRound != change.FromRound+1 {
		return false, RoundCertificate{}, ErrFutureRound
	}
	committee := e.committeeForHeightLocked(height)
	validator, ok := findValidator(committee, change.ValidatorID)
	if !ok {
		return false, RoundCertificate{}, ErrUnknownValidator
	}
	if !VerifySignature(validator.Algorithm, validator.PublicKey, roundChangeSigningMessage(change), change.Signature) {
		return false, RoundCertificate{}, ErrInvalidSignature
	}
	if change.LockedValueHash == "" {
		if change.LockProof != nil || change.LockedRound != 0 {
			return false, RoundCertificate{}, ErrInvalidRoundChange
		}
	} else {
		if change.LockProof == nil {
			return false, RoundCertificate{}, ErrInvalidRoundChange
		}
		if err := verifyLockProof(change, validator, *change.LockProof); err != nil {
			return false, RoundCertificate{}, err
		}
		if change.LockedRound > change.FromRound {
			return false, RoundCertificate{}, ErrInvalidRoundChange
		}
	}
	if lock, ok := e.validatorLocks[change.ValidatorID]; ok {
		if change.LockedValueHash != lock.ValueHash || change.LockedRound != lock.Round {
			return false, RoundCertificate{}, ErrLockedValue
		}
	} else if change.LockProof != nil {
		e.validatorLocks[change.ValidatorID] = ValidatorLock{ValidatorID: change.ValidatorID, Round: change.LockedRound, ValueHash: change.LockedValueHash, Proof: *change.LockProof}
	}
	bucket := e.roundChanges[change.ToRound]
	if bucket == nil {
		bucket = map[string]RoundChange{}
		e.roundChanges[change.ToRound] = bucket
	}
	if prior, ok := bucket[change.ValidatorID]; ok {
		if prior.Signature == change.Signature {
			return false, RoundCertificate{}, ErrInvalidRoundChange
		}
		return false, RoundCertificate{}, ErrEquivocation
	}
	bucket[change.ValidatorID] = change
	quorum := QuorumFor(e.genesis.Config, len(committee))
	if len(bucket) < quorum {
		return false, RoundCertificate{}, nil
	}
	changes := sortedRoundChanges(bucket)
	threshold := quorumIntersectionThreshold(len(committee), quorum)
	lockedRound, lockedValue, err := chooseCertificateLock(changes, threshold)
	if err != nil {
		return false, RoundCertificate{}, err
	}
	cert := RoundCertificate{NetworkID: e.genesis.NetworkID, Height: height, Round: change.ToRound, CommitteeHash: CommitteeHash(committee), Quorum: quorum, LockedRound: lockedRound, LockedValueHash: lockedValue, Changes: changes}
	if err := e.verifyRoundCertificateLocked(cert); err != nil {
		return false, RoundCertificate{}, err
	}
	e.roundCertificates[cert.Round] = cert
	e.recordCertificateLocksLocked(cert)
	e.currentRound = cert.Round
	return true, cert, nil
}

func (e *Engine) ApplyRoundCertificate(cert RoundCertificate) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.applyRoundCertificateLocked(cert)
}

func (e *Engine) applyRoundCertificateLocked(cert RoundCertificate) error {
	if cert.Round < e.currentRound {
		return ErrStaleRound
	}
	if cert.Round == e.currentRound {
		if existing, ok := e.roundCertificates[cert.Round]; ok && existing.CommitteeHash == cert.CommitteeHash && existing.LockedValueHash == cert.LockedValueHash {
			return nil
		}
	}
	if cert.Round != e.currentRound+1 {
		return ErrFutureRound
	}
	if err := e.verifyRoundCertificateLocked(cert); err != nil {
		return err
	}
	e.roundCertificates[cert.Round] = cert
	e.recordCertificateLocksLocked(cert)
	e.currentRound = cert.Round
	return nil
}

func (e *Engine) verifyRoundCertificateLocked(cert RoundCertificate) error {
	height := uint64(len(e.finalized)) + 1
	if cert.NetworkID != e.genesis.NetworkID || cert.Height != height || cert.Round == 0 {
		return ErrInvalidRoundChange
	}
	committee := e.committeeForHeightLocked(height)
	if cert.CommitteeHash != CommitteeHash(committee) {
		return ErrInvalidCommittee
	}
	quorum := QuorumFor(e.genesis.Config, len(committee))
	if cert.Quorum != quorum || len(cert.Changes) < quorum {
		return ErrInsufficientQuorum
	}
	seen := map[string]struct{}{}
	valid := make([]RoundChange, 0, len(cert.Changes))
	for _, change := range cert.Changes {
		if change.NetworkID != e.genesis.NetworkID || change.Height != height || change.ToRound != cert.Round || change.FromRound+1 != change.ToRound {
			return ErrInvalidRoundChange
		}
		if _, dup := seen[change.ValidatorID]; dup {
			return ErrInvalidRoundChange
		}
		validator, ok := findValidator(committee, change.ValidatorID)
		if !ok {
			return ErrUnknownValidator
		}
		if !VerifySignature(validator.Algorithm, validator.PublicKey, roundChangeSigningMessage(change), change.Signature) {
			return ErrInvalidSignature
		}
		if change.LockedValueHash == "" {
			if change.LockProof != nil || change.LockedRound != 0 {
				return ErrInvalidRoundChange
			}
		} else {
			if change.LockProof == nil || change.LockedRound > change.FromRound {
				return ErrInvalidRoundChange
			}
			if err := verifyLockProof(change, validator, *change.LockProof); err != nil {
				return err
			}
		}
		seen[change.ValidatorID] = struct{}{}
		valid = append(valid, change)
	}
	threshold := quorumIntersectionThreshold(len(committee), quorum)
	lockedRound, lockedValue, err := chooseCertificateLock(valid, threshold)
	if err != nil {
		return err
	}
	if cert.LockedRound != lockedRound || cert.LockedValueHash != lockedValue {
		return ErrInvalidRoundChange
	}
	return nil
}

func (e *Engine) verifyValidatorLockLocked(lock ValidatorLock, height uint64, committee []Validator) error {
	if lock.ValidatorID == "" || lock.ValueHash == "" {
		return ErrLockedValue
	}
	validator, ok := findValidator(committee, lock.ValidatorID)
	if !ok {
		return ErrUnknownValidator
	}
	change := RoundChange{NetworkID: e.genesis.NetworkID, Height: height, FromRound: lock.Round, ToRound: lock.Round + 1, ValidatorID: lock.ValidatorID, LockedRound: lock.Round, LockedValueHash: lock.ValueHash, LockProof: &lock.Proof}
	if err := verifyLockProof(change, validator, lock.Proof); err != nil {
		return err
	}
	return nil
}

func (e *Engine) recordCertificateLocksLocked(cert RoundCertificate) {
	for _, change := range cert.Changes {
		if change.LockProof == nil || change.LockedValueHash == "" {
			continue
		}
		prior, ok := e.validatorLocks[change.ValidatorID]
		if !ok || change.LockedRound >= prior.Round {
			e.validatorLocks[change.ValidatorID] = ValidatorLock{ValidatorID: change.ValidatorID, Round: change.LockedRound, ValueHash: change.LockedValueHash, Proof: *change.LockProof}
		}
	}
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
	e.resetRoundLocked()
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
	if b.Round == 0 {
		if b.RoundCertificate != nil {
			return ErrInvalidRoundChange
		}
	} else {
		if b.RoundCertificate == nil || b.RoundCertificate.Round != b.Round {
			return ErrInvalidRoundChange
		}
		if err := e.verifyRoundCertificateLocked(*b.RoundCertificate); err != nil {
			return err
		}
		if b.RoundCertificate.LockedValueHash != "" && blockValueHash(b) != b.RoundCertificate.LockedValueHash {
			return ErrLockedValue
		}
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
	committee := e.committeeForHeightLocked(b.Height)
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
	valueHash := blockValueHash(b)
	for _, vote := range f.Certificate.Votes {
		if vote.BlockHash != b.Hash || vote.Height != b.Height || vote.Round != b.Round || vote.Decision != "commit" {
			continue
		}
		if b.Round > 0 && vote.ValueHash == "" {
			continue
		}
		if vote.ValueHash != "" && vote.ValueHash != valueHash {
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

func (e *Engine) roundProgressLocked() RoundProgress {
	height := uint64(len(e.finalized)) + 1
	locks := make([]ValidatorLock, 0, len(e.validatorLocks))
	for _, lock := range e.validatorLocks {
		locks = append(locks, lock)
	}
	sort.Slice(locks, func(i, j int) bool { return locks[i].ValidatorID < locks[j].ValidatorID })
	certs := make([]RoundCertificate, 0, len(e.roundCertificates))
	for _, cert := range e.roundCertificates {
		certs = append(certs, cert)
	}
	sort.Slice(certs, func(i, j int) bool { return certs[i].Round < certs[j].Round })
	return RoundProgress{Height: height, Round: e.currentRound, Locks: locks, Certificates: certs}
}

func (e *Engine) resetRoundLocked() {
	e.proposals = map[string]Proposal{}
	e.votes = map[string]map[string]Vote{}
	e.seenVotes = map[string]string{}
	e.currentRound = 0
	e.roundChanges = map[uint32]map[string]RoundChange{}
	e.roundCertificates = map[uint32]RoundCertificate{}
	e.validatorLocks = map[string]ValidatorLock{}
	e.proposalSlots = map[string]string{}
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
