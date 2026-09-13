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
	ConsensusStateVersion = 1
	FeePoolAccount        = "system:pending-fees"
)

type StateAccount struct {
	Account      string `json:"account"`
	BalanceAtoms int64  `json:"balanceAtoms"`
	Nonce        uint64 `json:"nonce"`
}

type ConsensusSnapshot struct {
	Version             int            `json:"version"`
	NetworkID           string         `json:"networkId"`
	CarrotPolicyHash    string         `json:"carrotPolicyHash"`
	Height              uint64         `json:"height"`
	SettledRewardHeight uint64         `json:"settledRewardHeight"`
	Accounts            []StateAccount `json:"accounts"`
}

// ConsensusState is the v0.12 replicated TEST-CARROT state machine. Rewards
// earned by finalizing height H are deterministically settled in block H+1,
// when H's finality certificate is already canonical and known to every node.
type ConsensusState struct {
	mu                  sync.RWMutex
	networkID           string
	policy               carrot.Policy
	balances             map[string]int64
	nonces               map[string]uint64
	rewardAddressByID    map[string]string
	height               uint64
	settledRewardHeight  uint64
}

func NewConsensusState(networkID string, policy carrot.Policy, validators []testnet.Validator) (*ConsensusState, error) {
	if networkID == "" { return nil, errors.New("network id required") }
	if err := policy.Validate(); err != nil { return nil, err }
	s := &ConsensusState{networkID: networkID, policy: policy, balances: map[string]int64{}, nonces: map[string]uint64{}, rewardAddressByID: map[string]string{}}
	for _, a := range policy.Allocations { s.balances[a.Account] = a.Atoms }
	for _, v := range validators { s.rewardAddressByID[v.ID] = rewardAddress(v) }
	if err := s.validateConservationLocked(); err != nil { return nil, err }
	return s, nil
}

func rewardAddress(v testnet.Validator) string {
	if ValidAddress(v.RewardAddress) { return v.RewardAddress }
	return "validator:" + v.ID
}

func (s *ConsensusState) Root() string { s.mu.RLock(); defer s.mu.RUnlock(); return s.rootLocked() }
func (s *ConsensusState) Preview(t testnet.Transition) (string, error) {
	s.mu.RLock(); clone := s.cloneLocked(); s.mu.RUnlock()
	if err := clone.apply(t); err != nil { return "", err }
	return clone.rootLocked(), nil
}
func (s *ConsensusState) Commit(t testnet.Transition) (string, error) {
	s.mu.Lock(); defer s.mu.Unlock()
	clone := s.cloneLocked(); if err := clone.apply(t); err != nil { return "", err }
	s.balances = clone.balances; s.nonces = clone.nonces; s.height = clone.height; s.settledRewardHeight = clone.settledRewardHeight
	return s.rootLocked(), nil
}
func (s *ConsensusState) Balance(account string) int64 { s.mu.RLock(); defer s.mu.RUnlock(); return s.balances[account] }
func (s *ConsensusState) Nonce(account string) uint64 { s.mu.RLock(); defer s.mu.RUnlock(); return s.nonces[account] }
func (s *ConsensusState) Height() uint64 { s.mu.RLock(); defer s.mu.RUnlock(); return s.height }
func (s *ConsensusState) SettledRewardHeight() uint64 { s.mu.RLock(); defer s.mu.RUnlock(); return s.settledRewardHeight }
func (s *ConsensusState) Snapshot() ConsensusSnapshot { s.mu.RLock(); defer s.mu.RUnlock(); return s.snapshotLocked() }
func (s *ConsensusState) ValidateConservation() error { s.mu.RLock(); defer s.mu.RUnlock(); return s.validateConservationLocked() }

func (s *ConsensusState) SupplyReport() carrot.SupplyReport {
	s.mu.RLock(); defer s.mu.RUnlock()
	founderUnlocked := carrot.FounderUnlockedAt(s.height); founderLocked := carrot.FounderAllocationAtoms-founderUnlocked
	participationRemaining := s.balances[carrot.AccountParticipation]; participationReleased := carrot.ParticipationReserveAtoms-participationRemaining
	treasuryReserved := s.balances[carrot.AccountEcosystem]+s.balances[carrot.AccountCommunity]+s.balances[carrot.AccountSecurity]
	circulating := carrot.MaxSupplyAtoms-founderLocked-participationRemaining-treasuryReserved; if circulating<0 { circulating=0 }
	return carrot.SupplyReport{Height:s.height,MaxSupplyAtoms:carrot.MaxSupplyAtoms,MaxSupplyCARROT:carrot.FormatAtoms(carrot.MaxSupplyAtoms),FounderUnlockedAtoms:founderUnlocked,FounderLockedAtoms:founderLocked,ParticipationReleasedAtoms:participationReleased,ParticipationRemainingAtoms:participationRemaining,TreasuryReservedAtoms:treasuryReserved,CirculatingAtoms:circulating,CirculatingCARROT:carrot.FormatAtoms(circulating)}
}

func (s *ConsensusState) cloneLocked() *ConsensusState {
	c := &ConsensusState{networkID:s.networkID,policy:s.policy,balances:map[string]int64{},nonces:map[string]uint64{},rewardAddressByID:map[string]string{},height:s.height,settledRewardHeight:s.settledRewardHeight}
	for k,v := range s.balances { c.balances[k]=v }; for k,v := range s.nonces { c.nonces[k]=v }; for k,v := range s.rewardAddressByID { c.rewardAddressByID[k]=v }
	return c
}
func (s *ConsensusState) apply(t testnet.Transition) error {
	if t.NetworkID != s.networkID { return testnet.ErrWrongNetwork }
	if t.Height != s.height+1 { return testnet.ErrInvalidHeight }
	if t.Height > 1 {
		if t.PreviousFinalized == nil || t.PreviousFinalized.Block.Height != t.Height-1 { return errors.New("previous finality certificate required") }
		if err := s.settlePrevious(t.Height-1, t.PreviousFinalized.Certificate); err != nil { return err }
	}
	seen := map[string]struct{}{}
	for _, op := range t.Operations {
		switch op.Type {
		case "noop":
			continue
		case TransactionOperationType:
			tx, err := ParseTransactionOperation(op); if err != nil { return err }
			if _, ok := seen[tx.ID]; ok { return ErrDuplicateTransaction }; seen[tx.ID]=struct{}{}
			if err := s.applyTransaction(tx,t.Height); err != nil { return err }
		default:
			return fmt.Errorf("unsupported consensus operation %q",op.Type)
		}
	}
	s.height=t.Height
	return s.validateConservationLocked()
}
func (s *ConsensusState) settlePrevious(rewardHeight uint64, cert testnet.FinalityCertificate) error {
	if s.settledRewardHeight+1 != rewardHeight { return fmt.Errorf("reward settlement gap: have %d want %d",s.settledRewardHeight,rewardHeight) }
	ids := make([]string,0,len(cert.Votes)); seen:=map[string]struct{}{}
	for _,v := range cert.Votes { if _,ok:=seen[v.ValidatorID];!ok { seen[v.ValidatorID]=struct{}{}; ids=append(ids,v.ValidatorID) } }
	sort.Strings(ids); if len(ids)==0 { return errors.New("reward settlement requires finality signers") }
	reward := carrot.RewardAtHeight(rewardHeight); if reward > s.balances[carrot.AccountParticipation] { return errors.New("participation reserve exhausted") }
	s.balances[carrot.AccountParticipation]-=reward; s.distributeToValidators(reward,ids)
	fees:=s.balances[FeePoolAccount]; if fees>0 { s.balances[FeePoolAccount]=0; s.distributeToValidators(fees,ids) }
	s.settledRewardHeight=rewardHeight; return nil
}
func (s *ConsensusState) distributeToValidators(amount int64, ids []string) {
	if amount<=0 || len(ids)==0 { return }; base:=amount/int64(len(ids)); rem:=int(amount%int64(len(ids)))
	for i,id:=range ids { share:=base; if i<rem { share++ }; addr:=s.rewardAddressByID[id]; if addr=="" { addr="validator:"+id }; s.balances[addr]+=share }
}
func (s *ConsensusState) applyTransaction(tx SignedTransaction,height uint64) error {
	expected:=s.nonces[tx.From]; if err:=ValidateTransaction(tx,s.networkID,height,expected);err!=nil{return err}; total:=tx.AmountAtoms+tx.FeeAtoms; if total<tx.AmountAtoms || s.balances[tx.From]<total { return carrot.ErrInsufficientBalance }
	s.balances[tx.From]-=total; s.balances[tx.To]+=tx.AmountAtoms; if tx.FeeAtoms>0 { s.balances[FeePoolAccount]+=tx.FeeAtoms }; s.nonces[tx.From]=expected+1; return nil
}
func (s *ConsensusState) snapshotLocked() ConsensusSnapshot {
	keys:=map[string]struct{}{}; for k:=range s.balances{keys[k]=struct{}{}};for k:=range s.nonces{keys[k]=struct{}{}}
	names:=make([]string,0,len(keys));for k:=range keys{names=append(names,k)};sort.Strings(names);accounts:=make([]StateAccount,0,len(names));for _,k:=range names{accounts=append(accounts,StateAccount{Account:k,BalanceAtoms:s.balances[k],Nonce:s.nonces[k]})}
	return ConsensusSnapshot{Version:ConsensusStateVersion,NetworkID:s.networkID,CarrotPolicyHash:s.policy.Hash(),Height:s.height,SettledRewardHeight:s.settledRewardHeight,Accounts:accounts}
}
func (s *ConsensusState) rootLocked() string { raw,_:=json.Marshal(s.snapshotLocked());sum:=sha256.Sum256(raw);return hex.EncodeToString(sum[:]) }
func (s *ConsensusState) validateConservationLocked() error { var total int64; for account,b:=range s.balances { if b<0{return fmt.Errorf("negative balance %s",account)}; total+=b }; if total!=carrot.MaxSupplyAtoms{return fmt.Errorf("TEST-CARROT supply conservation failed: got %d want %d",total,carrot.MaxSupplyAtoms)};return nil }
