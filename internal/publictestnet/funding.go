package publictestnet

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"github.com/chrisbirster/battlebunnywealth/internal/carrot"
)

const TestFundingPolicyVersion = 1

var ErrInvalidTestFundingPolicy = errors.New("invalid TEST-CARROT funding policy")

type TestFundingPolicy struct {
	Version                int           `json:"version"`
	Asset                  string        `json:"asset"`
	EconomicValue          bool          `json:"economicValue"`
	Mode                   string        `json:"mode"`
	AutomaticFaucetEnabled bool          `json:"automaticFaucetEnabled"`
	TreasurySpending       bool          `json:"treasurySpending"`
	MaxGrantAtoms          int64         `json:"maxGrantAtoms"`
	Window                 time.Duration `json:"window"`
	MaxGrantsPerAddress    int           `json:"maxGrantsPerAddress"`
	OperatorSource         string        `json:"operatorSource"`
	Hash                   string        `json:"hash"`
}

func DefaultTestFundingPolicy() TestFundingPolicy {
	p := TestFundingPolicy{
		Version:                TestFundingPolicyVersion,
		Asset:                  "TEST-CARROT",
		EconomicValue:          false,
		Mode:                   "manual-existing-balance-transfer",
		AutomaticFaucetEnabled: false,
		TreasurySpending:       false,
		MaxGrantAtoms:          100 * carrot.AtomsPerCARROT,
		Window:                 24 * time.Hour,
		MaxGrantsPerAddress:    1,
		OperatorSource:         "operator-owned-test-wallet",
	}
	p.Hash = testFundingPolicyHash(p)
	return p
}

func (p TestFundingPolicy) Validate() error {
	if p.Version != TestFundingPolicyVersion || p.Asset != "TEST-CARROT" || p.EconomicValue || p.Mode != "manual-existing-balance-transfer" {
		return ErrInvalidTestFundingPolicy
	}
	if p.AutomaticFaucetEnabled || p.TreasurySpending || p.MaxGrantAtoms <= 0 || p.Window <= 0 || p.MaxGrantsPerAddress <= 0 || p.OperatorSource == "" {
		return ErrInvalidTestFundingPolicy
	}
	if p.Hash != testFundingPolicyHash(p) {
		return ErrInvalidTestFundingPolicy
	}
	return nil
}

func testFundingPolicyHash(p TestFundingPolicy) string {
	p.Hash = ""
	raw, _ := json.Marshal(p)
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}
