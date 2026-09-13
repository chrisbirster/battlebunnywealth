package publictestnet

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/chrisbirster/battlebunnywealth/internal/carrot"
	"github.com/chrisbirster/battlebunnywealth/internal/testnet"
)

const (
	ConsensusStateVersion = 2
	FeePoolAccount         = "system:pending-fees"
)

type StateAccount struct {
	Account      string `json:"account"`
	BalanceAtoms int64  `json:"balanceAtoms"`
	Nonce        uint64 `json:"nonce"`
}

type ValidatorSetActivation struct {
	ActivationHeight uint64              `json:"activationHeight"`
	PlanHash         string              `json:"planHash"`
	Validators       []testnet.Validator `json:"validators"`
}

type ProtocolActivation struct {
	ActivationHeight uint64 `json:"activationHeight"`
	PlanHash         string `json:"planHash"`
	Version          int    `json:"version"`
	MinimumSoftware  string `json:"minimumSoftware,omitempty"`
}

type ConsensusSnapshot struct {
	Version               int                      `json:"version"`
	NetworkID             string                   `json:"networkId"`
	CarrotPolicyHash      string                   `json:"carrotPolicyHash"`
	Height                uint64                   `json:"height"`
	SettledRewardHeight   uint64                   `json:"settledRewardHeight"`
	ActiveProtocolVersion int                      `json:"activeProtocolVersion"`
	PendingUpgrade        *UpgradePlan             `json:"pendingUpgrade,omitempty"`
	ProtocolHistory       []ProtocolActivation     `json:"protocolHistory"`
	PendingValidatorSet   *ValidatorSetPlan        `json:"pendingValidatorSet,omitempty"`
	ValidatorSetHistory   []ValidatorSetActivation `json:"validatorSetHistory"`
	Accounts              []StateAccount           `json:"accounts"`
}

type TransitionStatus struct {
	ActiveProtocolVersion int               `json:"activeProtocolVersion"`
	ActiveValidatorCount  int               `json:"activeValidatorCount"`
	PendingUpgrade        *UpgradePlan      `json:"pendingUpgrade,omitempty"`
	PendingValidatorSet   *ValidatorSetPlan `json:"pendingValidatorSet,omitempty"`
}

type ConsensusState struct {
	mu                    sync.RWMutex
	genesis               testnet.Genesis
	networkID             string
	policy                carrot.Policy
	balances              map[string]int64
	nonces                map[string]uint64
	rewardAddressByID     map[string]string
	height                uint64
	settledRewardHeight   uint64
	activeValidators      []testnet.Validator
	validatorSetHistory   []ValidatorSetActivation
	pendingValidatorSet   *ValidatorSetPlan
	activeProtocolVersion int
	protocolHistory       []ProtocolActivation
	pendingUpgrade        *UpgradePlan
}

func NewConsensusState(genesis testnet.Genesis, policy carrot.Policy) (*ConsensusState, error) {
	if err := genesis.Validate(); err != nil {
		return nil, err
	}
	if err := policy.Validate(); err != nil {
		return nil, err
	}
	if genesis.CarrotPolicyHash != policy.Hash() {
		return nil, errors.New("CARROT policy hash mismatch")
	}
	validators := append([]testnet.Validator(nil), genesis.Validators...)
	s := &ConsensusState{
		genesis:               genesis,
		networkID:             genesis.NetworkID,
		policy:                policy,
		balances:              map[string]int64{},
		nonces:                map[string]uint64{},
		rewardAddressByID:     map[string]string{},
		activeValidators:      validators,
		activeProtocolVersion: genesis.Version,
		validatorSetHistory: []ValidatorSetActivation{{
			ActivationHeight: 1,
			PlanHash:         "genesis",
			Validators:       append([]testnet.Validator(nil), validators...),
		}},
		protocolHistory: []ProtocolActivation{{
			ActivationHeight: 1,
			PlanHash:         "genesis",
			Version:          genesis.Version,
		}},
	}
	for _, a := range policy.Allocations {
		s.balances[a.Account] = a.Atoms
	}
	for _, v := range genesis.Validators {
		s.rewardAddressByID[v.ID] = rewardAddress(v)
	}
	if err := s.validateConservationLocked(); err != nil {
		return nil, err
	}
	return s, nil
}

func rewardAddress(v testnet.Validator) string {
	if ValidAddress(v.RewardAddress) {
		return v.RewardAddress
	}
	return "validator:" + v.ID
}

func (s *ConsensusState) Root() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.rootLocked()
}

func (s *ConsensusState) Preview(t testnet.Transition) (string, error) {
	s.mu.RLock()
	c := s.cloneLocked()
	s.mu.RUnlock()
	if err := c.apply(t); err != nil {
		return "", err
	}
	return c.rootLocked(), nil
}

func (s *ConsensusState) Commit(t testnet.Transition) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c := s.cloneLocked()
	if err := c.apply(t); err != nil {
		return "", err
	}
	s.balances = c.balances
	s.nonces = c.nonces
	s.rewardAddressByID = c.rewardAddressByID
	s.height = c.height
	s.settledRewardHeight = c.settledRewardHeight
	s.activeValidators = c.activeValidators
	s.validatorSetHistory = c.validatorSetHistory
	s.pendingValidatorSet = c.pendingValidatorSet
	s.activeProtocolVersion = c.activeProtocolVersion
	s.protocolHistory = c.protocolHistory
	s.pendingUpgrade = c.pendingUpgrade
	return s.rootLocked(), nil
}

// ValidatorsForHeight implements testnet.ConsensusRules. It returns the
// historical or already-committed scheduled set that controls the requested
// height. A plan becomes consensus-effective at its activation height even
// before that height's block is committed, because the prior block already
// finalized the plan.
func (s *ConsensusState) ValidatorsForHeight(height uint64) []testnet.Validator {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]testnet.Validator(nil), s.validatorsForHeightLocked(height)...)
}

// ProtocolVersionForHeight implements testnet.ConsensusRules.
func (s *ConsensusState) ProtocolVersionForHeight(height uint64) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.protocolVersionForHeightLocked(height)
}

func (s *ConsensusState) Balance(a string) int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.balances[a]
}

func (s *ConsensusState) Nonce(a string) uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.nonces[a]
}

func (s *ConsensusState) Height() uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.height
}

func (s *ConsensusState) SettledRewardHeight() uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.settledRewardHeight
}

func (s *ConsensusState) TransitionStatus() TransitionStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return TransitionStatus{
		ActiveProtocolVersion: s.activeProtocolVersion,
		ActiveValidatorCount:  len(s.activeValidators),
		PendingUpgrade:        cloneUpgradePlan(s.pendingUpgrade),
		PendingValidatorSet:   cloneValidatorSetPlan(s.pendingValidatorSet),
	}
}

func (s *ConsensusState) Snapshot() ConsensusSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.snapshotLocked()
}

func (s *ConsensusState) ValidateConservation() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.validateConservationLocked()
}

func (s *ConsensusState) SupplyReport() carrot.SupplyReport {
	s.mu.RLock()
	defer s.mu.RUnlock()
	fu := carrot.FounderUnlockedAt(s.height)
	fl := carrot.FounderAllocationAtoms - fu
	pr := s.balances[carrot.AccountParticipation]
	released := carrot.ParticipationReserveAtoms - pr
	tr := s.balances[carrot.AccountEcosystem] + s.balances[carrot.AccountCommunity] + s.balances[carrot.AccountSecurity]
	circ := carrot.MaxSupplyAtoms - fl - pr - tr
	if circ < 0 {
		circ = 0
	}
	return carrot.SupplyReport{
		Height:                      s.height,
		MaxSupplyAtoms:              carrot.MaxSupplyAtoms,
		MaxSupplyCARROT:             carrot.FormatAtoms(carrot.MaxSupplyAtoms),
		FounderUnlockedAtoms:        fu,
		FounderLockedAtoms:          fl,
		ParticipationReleasedAtoms:  released,
		ParticipationRemainingAtoms: pr,
		TreasuryReservedAtoms:       tr,
		CirculatingAtoms:            circ,
		CirculatingCARROT:           carrot.FormatAtoms(circ),
	}
}

func (s *ConsensusState) cloneLocked() *ConsensusState {
	c := &ConsensusState{
		genesis:               s.genesis,
		networkID:             s.networkID,
		policy:                s.policy,
		balances:              map[string]int64{},
		nonces:                map[string]uint64{},
		rewardAddressByID:     map[string]string{},
		height:                s.height,
		settledRewardHeight:   s.settledRewardHeight,
		activeValidators:      append([]testnet.Validator(nil), s.activeValidators...),
		validatorSetHistory:   cloneValidatorHistory(s.validatorSetHistory),
		pendingValidatorSet:   cloneValidatorSetPlan(s.pendingValidatorSet),
		activeProtocolVersion: s.activeProtocolVersion,
		protocolHistory:       append([]ProtocolActivation(nil), s.protocolHistory...),
		pendingUpgrade:        cloneUpgradePlan(s.pendingUpgrade),
	}
	for k, v := range s.balances {
		c.balances[k] = v
	}
	for k, v := range s.nonces {
		c.nonces[k] = v
	}
	for k, v := range s.rewardAddressByID {
		c.rewardAddressByID[k] = v
	}
	return c
}

func cloneValidatorHistory(in []ValidatorSetActivation) []ValidatorSetActivation {
	out := make([]ValidatorSetActivation, len(in))
	for i, h := range in {
		out[i] = h
		out[i].Validators = append([]testnet.Validator(nil), h.Validators...)
	}
	return out
}

func cloneValidatorSetPlan(in *ValidatorSetPlan) *ValidatorSetPlan {
	if in == nil {
		return nil
	}
	p := *in
	p.Validators = append([]testnet.Validator(nil), in.Validators...)
	return &p
}

func cloneUpgradePlan(in *UpgradePlan) *UpgradePlan {
	if in == nil {
		return nil
	}
	p := *in
	return &p
}

func (s *ConsensusState) validatorsForHeightLocked(height uint64) []testnet.Validator {
	var selected []testnet.Validator
	for _, activation := range s.validatorSetHistory {
		if activation.ActivationHeight <= height {
			selected = activation.Validators
		} else {
			break
		}
	}
	if s.pendingValidatorSet != nil && s.pendingValidatorSet.ActivationHeight <= height {
		selected = s.pendingValidatorSet.Validators
	}
	if len(selected) == 0 {
		selected = s.genesis.Validators
	}
	return selected
}

func (s *ConsensusState) protocolVersionForHeightLocked(height uint64) int {
	version := s.genesis.Version
	for _, activation := range s.protocolHistory {
		if activation.ActivationHeight <= height {
			version = activation.Version
		} else {
			break
		}
	}
	if s.pendingUpgrade != nil && s.pendingUpgrade.ActivationHeight <= height {
		version = s.pendingUpgrade.ToVersion
	}
	return version
}

func (s *ConsensusState) activateScheduledLocked(height uint64) error {
	if s.pendingValidatorSet != nil {
		if s.pendingValidatorSet.ActivationHeight < height {
			return errors.New("missed validator-set activation height")
		}
		if s.pendingValidatorSet.ActivationHeight == height {
			p := s.pendingValidatorSet
			s.activeValidators = append([]testnet.Validator(nil), p.Validators...)
			s.validatorSetHistory = append(s.validatorSetHistory, ValidatorSetActivation{
				ActivationHeight: p.ActivationHeight,
				PlanHash:         p.Hash,
				Validators:       append([]testnet.Validator(nil), p.Validators...),
			})
			for _, v := range p.Validators {
				s.rewardAddressByID[v.ID] = rewardAddress(v)
			}
			s.pendingValidatorSet = nil
		}
	}
	if s.pendingUpgrade != nil {
		if s.pendingUpgrade.ActivationHeight < height {
			return errors.New("missed protocol-upgrade activation height")
		}
		if s.pendingUpgrade.ActivationHeight == height {
			p := s.pendingUpgrade
			if p.ToVersion > testnet.MaxSupportedProtocolVersion {
				return fmt.Errorf("protocol version %d is not supported by this binary", p.ToVersion)
			}
			s.activeProtocolVersion = p.ToVersion
			s.protocolHistory = append(s.protocolHistory, ProtocolActivation{
				ActivationHeight: p.ActivationHeight,
				PlanHash:         p.Hash,
				Version:          p.ToVersion,
				MinimumSoftware:  p.MinimumSoftware,
			})
			s.pendingUpgrade = nil
		}
	}
	return nil
}

func (s *ConsensusState) apply(t testnet.Transition) error {
	if t.NetworkID != s.networkID {
		return testnet.ErrWrongNetwork
	}
	if t.Height != s.height+1 {
		return testnet.ErrInvalidHeight
	}
	if err := s.activateScheduledLocked(t.Height); err != nil {
		return err
	}

	var settlement *FinalitySettlement
	seenTx := map[string]struct{}{}
	seenValidatorCommitment := false
	seenUpgradeCommitment := false
	for _, op := range t.Operations {
		if op.Type == SettlementOperationType {
			x, err := ParseSettlementOperation(op)
			if err != nil {
				return err
			}
			if settlement != nil {
				return errors.New("duplicate finality settlement")
			}
			settlement = &x
		}
	}

	if t.Height == 1 {
		if settlement != nil {
			return errors.New("height 1 cannot settle prior finality")
		}
	} else {
		if settlement == nil {
			return errors.New("missing prior finality settlement")
		}
		if t.PreviousFinalized == nil {
			return errors.New("previous finalized block required")
		}
		if err := s.settlePrevious(t.Height-1, *t.PreviousFinalized, *settlement); err != nil {
			return err
		}
	}

	for _, op := range t.Operations {
		switch op.Type {
		case "noop", SettlementOperationType:
			continue
		case TransactionOperationType:
			tx, err := ParseTransactionOperation(op)
			if err != nil {
				return err
			}
			if _, ok := seenTx[tx.ID]; ok {
				return ErrDuplicateTransaction
			}
			seenTx[tx.ID] = struct{}{}
			if err := s.applyTransaction(tx, t.Height); err != nil {
				return err
			}
		case ValidatorSetCommitmentOperationType:
			if seenValidatorCommitment || s.pendingValidatorSet != nil {
				return errors.New("validator-set commitment already pending")
			}
			plan, err := ParseValidatorSetCommitmentOperation(op)
			if err != nil {
				return err
			}
			if err := plan.Validate(s.networkID, t.Height); err != nil {
				return err
			}
			for _, v := range plan.Validators {
				s.rewardAddressByID[v.ID] = rewardAddress(v)
			}
			s.pendingValidatorSet = cloneValidatorSetPlan(&plan)
			seenValidatorCommitment = true
		case UpgradeCommitmentOperationType:
			if seenUpgradeCommitment || s.pendingUpgrade != nil {
				return errors.New("protocol-upgrade commitment already pending")
			}
			plan, err := ParseUpgradeCommitmentOperation(op)
			if err != nil {
				return err
			}
			if err := plan.Validate(s.networkID, s.activeProtocolVersion, t.Height); err != nil {
				return err
			}
			if plan.ToVersion > testnet.MaxSupportedProtocolVersion {
				return fmt.Errorf("upgrade target %d exceeds supported version %d", plan.ToVersion, testnet.MaxSupportedProtocolVersion)
			}
			s.pendingUpgrade = cloneUpgradePlan(&plan)
			seenUpgradeCommitment = true
		default:
			return fmt.Errorf("unsupported consensus operation %q", op.Type)
		}
	}

	s.height = t.Height
	return s.validateConservationLocked()
}

func (s *ConsensusState) settlePrevious(rewardHeight uint64, previous testnet.FinalizedBlock, settlement FinalitySettlement) error {
	if s.settledRewardHeight+1 != rewardHeight {
		return fmt.Errorf("reward settlement gap: have %d want %d", s.settledRewardHeight, rewardHeight)
	}
	if previous.Block.Height != rewardHeight || settlement.RewardHeight != rewardHeight || settlement.BlockHash != previous.Block.Hash || settlement.Round != previous.Block.Round {
		return errors.New("settlement does not match previous block")
	}
	validators := s.validatorsForHeightLocked(rewardHeight)
	committee := testnet.SelectCommittee(validators, s.genesis.Config.CommitteeTarget, previous.Block.PreviousHash, previous.Block.Height, previous.Block.Round)
	quorum := testnet.QuorumFor(s.genesis.Config, len(committee))
	seen := map[string]struct{}{}
	ids := []string{}
	for _, vote := range settlement.Votes {
		if vote.NetworkID != s.networkID || vote.Height != previous.Block.Height || vote.Round != previous.Block.Round || vote.BlockHash != previous.Block.Hash {
			continue
		}
		if _, dup := seen[vote.ValidatorID]; dup {
			continue
		}
		var validator testnet.Validator
		found := false
		for _, v := range committee {
			if v.ID == vote.ValidatorID {
				validator = v
				found = true
				break
			}
		}
		if !found || !testnet.VerifyCommitVote(vote, validator) {
			continue
		}
		seen[vote.ValidatorID] = struct{}{}
		ids = append(ids, vote.ValidatorID)
	}
	if len(ids) < quorum {
		return testnet.ErrInsufficientQuorum
	}
	sort.Strings(ids)
	reward := carrot.RewardAtHeight(rewardHeight)
	if reward > s.balances[carrot.AccountParticipation] {
		return errors.New("participation reserve exhausted")
	}
	s.balances[carrot.AccountParticipation] -= reward
	s.distributeToValidators(reward, ids)
	fees := s.balances[FeePoolAccount]
	if fees > 0 {
		s.balances[FeePoolAccount] = 0
		s.distributeToValidators(fees, ids)
	}
	s.settledRewardHeight = rewardHeight
	return nil
}

func (s *ConsensusState) distributeToValidators(amount int64, ids []string) {
	if amount <= 0 || len(ids) == 0 {
		return
	}
	base := amount / int64(len(ids))
	rem := int(amount % int64(len(ids)))
	for i, id := range ids {
		share := base
		if i < rem {
			share++
		}
		addr := s.rewardAddressByID[id]
		if addr == "" {
			addr = "validator:" + id
		}
		s.balances[addr] += share
	}
}

func (s *ConsensusState) applyTransaction(tx SignedTransaction, height uint64) error {
	expected := s.nonces[tx.From]
	if err := ValidateTransaction(tx, s.networkID, height, expected); err != nil {
		return err
	}
	total := tx.AmountAtoms + tx.FeeAtoms
	if total < tx.AmountAtoms || s.balances[tx.From] < total {
		return carrot.ErrInsufficientBalance
	}
	s.balances[tx.From] -= total
	s.balances[tx.To] += tx.AmountAtoms
	if tx.FeeAtoms > 0 {
		s.balances[FeePoolAccount] += tx.FeeAtoms
	}
	s.nonces[tx.From] = expected + 1
	return nil
}

func (s *ConsensusState) snapshotLocked() ConsensusSnapshot {
	keys := map[string]struct{}{}
	for k := range s.balances {
		keys[k] = struct{}{}
	}
	for k := range s.nonces {
		keys[k] = struct{}{}
	}
	names := make([]string, 0, len(keys))
	for k := range keys {
		names = append(names, k)
	}
	sort.Strings(names)
	accounts := make([]StateAccount, 0, len(names))
	for _, k := range names {
		accounts = append(accounts, StateAccount{Account: k, BalanceAtoms: s.balances[k], Nonce: s.nonces[k]})
	}
	return ConsensusSnapshot{
		Version:               ConsensusStateVersion,
		NetworkID:             s.networkID,
		CarrotPolicyHash:      s.policy.Hash(),
		Height:                s.height,
		SettledRewardHeight:   s.settledRewardHeight,
		ActiveProtocolVersion: s.activeProtocolVersion,
		PendingUpgrade:        cloneUpgradePlan(s.pendingUpgrade),
		ProtocolHistory:       append([]ProtocolActivation(nil), s.protocolHistory...),
		PendingValidatorSet:   cloneValidatorSetPlan(s.pendingValidatorSet),
		ValidatorSetHistory:   cloneValidatorHistory(s.validatorSetHistory),
		Accounts:              accounts,
	}
}

func (s *ConsensusState) rootLocked() string {
	raw, _ := json.Marshal(s.snapshotLocked())
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func (s *ConsensusState) validateConservationLocked() error {
	var total int64
	for account, balance := range s.balances {
		if balance < 0 {
			return fmt.Errorf("negative balance %s", account)
		}
		total += balance
	}
	if total != carrot.MaxSupplyAtoms {
		return fmt.Errorf("TEST-CARROT supply conservation failed: got %d want %d", total, carrot.MaxSupplyAtoms)
	}
	return nil
}
