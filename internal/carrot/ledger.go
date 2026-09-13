package carrot

import (
	"errors"
	"fmt"
	"sort"
)

var (
	ErrInvalidAmount       = errors.New("invalid CARROT amount")
	ErrInsufficientBalance = errors.New("insufficient CARROT balance")
	ErrReservedAccount     = errors.New("CARROT reserve account cannot be spent by generic transfer")
	ErrFounderLocked       = errors.New("founder CARROT remains locked")
	ErrInvalidHeight       = errors.New("CARROT release height must advance exactly once")
	ErrNoRecipients        = errors.New("CARROT distribution requires recipients")
)

type Ledger struct {
	policy   Policy
	balances map[string]int64
	height   uint64
}

type SupplyReport struct {
	Height                      uint64 `json:"height"`
	MaxSupplyAtoms              int64  `json:"maxSupplyAtoms"`
	MaxSupplyCARROT             string `json:"maxSupplyCarrot"`
	FounderUnlockedAtoms        int64  `json:"founderUnlockedAtoms"`
	FounderLockedAtoms          int64  `json:"founderLockedAtoms"`
	ParticipationReleasedAtoms  int64  `json:"participationReleasedAtoms"`
	ParticipationRemainingAtoms int64  `json:"participationRemainingAtoms"`
	TreasuryReservedAtoms       int64  `json:"treasuryReservedAtoms"`
	CirculatingAtoms            int64  `json:"circulatingAtoms"`
	CirculatingCARROT           string `json:"circulatingCarrot"`
}

func NewLedger(policy Policy) (*Ledger, error) {
	if err := policy.Validate(); err != nil {
		return nil, err
	}
	balances := map[string]int64{}
	for _, a := range policy.Allocations {
		balances[a.Account] = a.Atoms
	}
	return &Ledger{policy: policy, balances: balances}, nil
}

func (l *Ledger) Height() uint64               { return l.height }
func (l *Ledger) Balance(account string) int64 { return l.balances[account] }
func (l *Ledger) PolicyHash() string           { return l.policy.Hash() }

func FounderUnlockedAt(height uint64) int64 {
	if height < FounderCliffEpochs {
		return 0
	}
	if height >= FounderVestingEpochs {
		return FounderAllocationAtoms
	}
	q := FounderAllocationAtoms / int64(FounderVestingEpochs)
	r := FounderAllocationAtoms % int64(FounderVestingEpochs)
	return q*int64(height) + r*int64(height)/int64(FounderVestingEpochs)
}

func EraBudget(era uint32) int64 {
	if era >= IssuanceEraCount {
		return 0
	}
	remaining := ParticipationReserveAtoms
	for i := uint32(0); i < era; i++ {
		remaining -= remaining / 2
	}
	if era == IssuanceEraCount-1 {
		return remaining
	}
	return remaining / 2
}

func RewardAtHeight(height uint64) int64 {
	if height == 0 {
		return 0
	}
	zero := height - 1
	era := uint32(zero / IssuanceEraEpochs)
	if era >= IssuanceEraCount {
		return 0
	}
	offset := zero % IssuanceEraEpochs
	budget := EraBudget(era)
	base := budget / int64(IssuanceEraEpochs)
	remainder := uint64(budget % int64(IssuanceEraEpochs))
	if offset < remainder {
		return base + 1
	}
	return base
}

func TotalScheduledParticipation() int64 {
	var total int64
	for era := uint32(0); era < IssuanceEraCount; era++ {
		total += EraBudget(era)
	}
	return total
}

func (l *Ledger) ReleaseParticipation(height uint64, recipients []string) (map[string]int64, error) {
	if height != l.height+1 {
		return nil, ErrInvalidHeight
	}
	unique := normalizeRecipients(recipients)
	if len(unique) == 0 {
		return nil, ErrNoRecipients
	}
	reward := RewardAtHeight(height)
	if reward > l.balances[AccountParticipation] {
		return nil, ErrInsufficientBalance
	}
	distributed := distribute(reward, unique)
	l.balances[AccountParticipation] -= reward
	for account, amount := range distributed {
		l.balances[account] += amount
	}
	l.height = height
	return distributed, nil
}

// Transfer models fee accounting only; wallet authentication and transaction
// admission live outside this package. v0.10 treasury spends remain disabled.
func (l *Ledger) Transfer(from, to string, amount, fee int64, feeRecipients []string) error {
	if amount <= 0 || fee < l.policy.Fees.MinimumFeeAtoms || fee < 0 || from == "" || to == "" || from == to {
		return ErrInvalidAmount
	}
	if from == AccountParticipation || from == AccountEcosystem || from == AccountCommunity || from == AccountSecurity {
		return ErrReservedAccount
	}
	if fee > 0 && len(normalizeRecipients(feeRecipients)) == 0 {
		return ErrNoRecipients
	}
	total := amount + fee
	if total < amount || l.balances[from] < total {
		return ErrInsufficientBalance
	}
	if from == AccountFounder {
		spent := FounderAllocationAtoms - l.balances[AccountFounder]
		spendable := FounderUnlockedAt(l.height) - spent
		if spendable < total {
			return ErrFounderLocked
		}
	}
	l.balances[from] -= total
	l.balances[to] += amount
	for account, share := range distribute(fee, normalizeRecipients(feeRecipients)) {
		l.balances[account] += share
	}
	return nil
}

func (l *Ledger) SupplyReport() SupplyReport {
	founderUnlocked := FounderUnlockedAt(l.height)
	founderLocked := FounderAllocationAtoms - founderUnlocked
	participationRemaining := l.balances[AccountParticipation]
	participationReleased := ParticipationReserveAtoms - participationRemaining
	treasuryReserved := l.balances[AccountEcosystem] + l.balances[AccountCommunity] + l.balances[AccountSecurity]
	circulating := MaxSupplyAtoms - founderLocked - participationRemaining - treasuryReserved
	if circulating < 0 {
		circulating = 0
	}
	return SupplyReport{
		Height:                      l.height,
		MaxSupplyAtoms:              MaxSupplyAtoms,
		MaxSupplyCARROT:             FormatAtoms(MaxSupplyAtoms),
		FounderUnlockedAtoms:        founderUnlocked,
		FounderLockedAtoms:          founderLocked,
		ParticipationReleasedAtoms:  participationReleased,
		ParticipationRemainingAtoms: participationRemaining,
		TreasuryReservedAtoms:       treasuryReserved,
		CirculatingAtoms:            circulating,
		CirculatingCARROT:           FormatAtoms(circulating),
	}
}

func (l *Ledger) ValidateConservation() error {
	var total int64
	for account, balance := range l.balances {
		if balance < 0 {
			return fmt.Errorf("negative balance %s", account)
		}
		total += balance
	}
	if total != MaxSupplyAtoms {
		return fmt.Errorf("CARROT supply conservation failed: got %d want %d", total, MaxSupplyAtoms)
	}
	return nil
}

func normalizeRecipients(recipients []string) []string {
	set := map[string]struct{}{}
	for _, r := range recipients {
		if r != "" {
			set[r] = struct{}{}
		}
	}
	out := make([]string, 0, len(set))
	for r := range set {
		out = append(out, r)
	}
	sort.Strings(out)
	return out
}

func distribute(amount int64, recipients []string) map[string]int64 {
	out := map[string]int64{}
	if amount <= 0 || len(recipients) == 0 {
		return out
	}
	base := amount / int64(len(recipients))
	remainder := int(amount % int64(len(recipients)))
	for i, r := range recipients {
		share := base
		if i < remainder {
			share++
		}
		out[r] = share
	}
	return out
}

func FormatAtoms(atoms int64) string {
	sign := ""
	if atoms < 0 {
		sign = "-"
		atoms = -atoms
	}
	whole := atoms / AtomsPerCARROT
	frac := atoms % AtomsPerCARROT
	return fmt.Sprintf("%s%d.%08d", sign, whole, frac)
}
