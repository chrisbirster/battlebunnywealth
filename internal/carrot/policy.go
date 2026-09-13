package carrot

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
)

const (
	ProtocolID            = "carrot/1"
	Symbol                = "CARROT"
	Decimals              = 8
	AtomsPerCARROT  int64 = 100_000_000
	MaxSupplyCARROT int64 = 21_000_000
	MaxSupplyAtoms  int64 = MaxSupplyCARROT * AtomsPerCARROT

	FounderBasisPoints       = 2_000
	ParticipationBasisPoints = 6_000
	EcosystemBasisPoints     = 1_000
	CommunityBasisPoints     = 500
	SecurityBasisPoints      = 500

	FounderAllocationAtoms    int64 = MaxSupplyAtoms / 10_000 * FounderBasisPoints
	ParticipationReserveAtoms int64 = MaxSupplyAtoms / 10_000 * ParticipationBasisPoints
	EcosystemReserveAtoms     int64 = MaxSupplyAtoms / 10_000 * EcosystemBasisPoints
	CommunityReserveAtoms     int64 = MaxSupplyAtoms / 10_000 * CommunityBasisPoints
	SecurityReserveAtoms      int64 = MaxSupplyAtoms / 10_000 * SecurityBasisPoints

	// Proof of Play currently targets ten-minute epochs. 52,560 epochs is one
	// 365-day year at that cadence; 210,240 epochs is four years.
	EpochSeconds         uint64 = 600
	EpochsPerYear        uint64 = 52_560
	FounderCliffEpochs   uint64 = EpochsPerYear
	FounderVestingEpochs uint64 = 4 * EpochsPerYear
	IssuanceEraEpochs    uint64 = 4 * EpochsPerYear
	IssuanceEraCount     uint32 = 32

	FounderCustodyThreshold         = 2
	FounderCustodyKeys              = 3
	TreasuryCustodyThreshold        = 3
	TreasuryCustodyKeys             = 5
	KeyRotationDelayEpochs   uint64 = 7 * 24 * 6
)

const (
	AccountFounder       = "allocation:founder"
	AccountParticipation = "reserve:participation"
	AccountEcosystem     = "treasury:ecosystem"
	AccountCommunity     = "treasury:community"
	AccountSecurity      = "treasury:security"
)

type Allocation struct {
	Account     string `json:"account"`
	Category    string `json:"category"`
	BasisPoints int    `json:"basisPoints"`
	Atoms       int64  `json:"atoms"`
	CARROT      string `json:"carrot"`
}

type CustodyPolicy struct {
	FounderThreshold        int    `json:"founderThreshold"`
	FounderKeyCount         int    `json:"founderKeyCount"`
	TreasuryThreshold       int    `json:"treasuryThreshold"`
	TreasuryKeyCount        int    `json:"treasuryKeyCount"`
	KeyRotationDelayEpochs  uint64 `json:"keyRotationDelayEpochs"`
	TreasurySpendingEnabled bool   `json:"treasurySpendingEnabled"`
}

type FeePolicy struct {
	Version         int    `json:"version"`
	MinimumFeeAtoms int64  `json:"minimumFeeAtoms"`
	BurnBasisPoints int    `json:"burnBasisPoints"`
	RecipientMode   string `json:"recipientMode"`
}

type Policy struct {
	Protocol             string        `json:"protocol"`
	Symbol               string        `json:"symbol"`
	Decimals             int           `json:"decimals"`
	MaxSupplyAtoms       int64         `json:"maxSupplyAtoms"`
	MaxSupplyCARROT      int64         `json:"maxSupplyCarrot"`
	EpochSeconds         uint64        `json:"epochSeconds"`
	FounderCliffEpochs   uint64        `json:"founderCliffEpochs"`
	FounderVestingEpochs uint64        `json:"founderVestingEpochs"`
	IssuanceEraEpochs    uint64        `json:"issuanceEraEpochs"`
	IssuanceEraCount     uint32        `json:"issuanceEraCount"`
	Allocations          []Allocation  `json:"allocations"`
	Fees                 FeePolicy     `json:"fees"`
	Custody              CustodyPolicy `json:"custody"`
}

func DefaultPolicy() Policy {
	return Policy{
		Protocol:             ProtocolID,
		Symbol:               Symbol,
		Decimals:             Decimals,
		MaxSupplyAtoms:       MaxSupplyAtoms,
		MaxSupplyCARROT:      MaxSupplyCARROT,
		EpochSeconds:         EpochSeconds,
		FounderCliffEpochs:   FounderCliffEpochs,
		FounderVestingEpochs: FounderVestingEpochs,
		IssuanceEraEpochs:    IssuanceEraEpochs,
		IssuanceEraCount:     IssuanceEraCount,
		Allocations: []Allocation{
			allocation(AccountFounder, "founder-admin", FounderBasisPoints, FounderAllocationAtoms),
			allocation(AccountParticipation, "proof-of-play-issuance", ParticipationBasisPoints, ParticipationReserveAtoms),
			allocation(AccountEcosystem, "ecosystem", EcosystemBasisPoints, EcosystemReserveAtoms),
			allocation(AccountCommunity, "community-treasury", CommunityBasisPoints, CommunityReserveAtoms),
			allocation(AccountSecurity, "security-public-goods", SecurityBasisPoints, SecurityReserveAtoms),
		},
		Fees: FeePolicy{Version: 1, MinimumFeeAtoms: 0, BurnBasisPoints: 0, RecipientMode: "finality-signers-equal"},
		Custody: CustodyPolicy{
			FounderThreshold:        FounderCustodyThreshold,
			FounderKeyCount:         FounderCustodyKeys,
			TreasuryThreshold:       TreasuryCustodyThreshold,
			TreasuryKeyCount:        TreasuryCustodyKeys,
			KeyRotationDelayEpochs:  KeyRotationDelayEpochs,
			TreasurySpendingEnabled: false,
		},
	}
}

func allocation(account, category string, bps int, atoms int64) Allocation {
	return Allocation{Account: account, Category: category, BasisPoints: bps, Atoms: atoms, CARROT: FormatAtoms(atoms)}
}

func (p Policy) Validate() error {
	if p.Protocol != ProtocolID || p.Symbol != Symbol || p.Decimals != Decimals {
		return errors.New("unexpected CARROT protocol identity")
	}
	if p.MaxSupplyAtoms != MaxSupplyAtoms || p.MaxSupplyCARROT != MaxSupplyCARROT {
		return errors.New("unexpected CARROT maximum supply")
	}
	if p.EpochSeconds != EpochSeconds || p.IssuanceEraEpochs != IssuanceEraEpochs || p.IssuanceEraCount != IssuanceEraCount {
		return errors.New("unexpected CARROT issuance cadence")
	}
	if p.FounderCliffEpochs != FounderCliffEpochs || p.FounderVestingEpochs != FounderVestingEpochs || p.FounderVestingEpochs <= p.FounderCliffEpochs {
		return errors.New("unexpected founder vesting policy")
	}
	if len(p.Allocations) != 5 {
		return errors.New("CARROT requires five fixed allocation buckets")
	}
	var atoms int64
	var bps int
	seen := map[string]struct{}{}
	for _, a := range p.Allocations {
		if a.Account == "" || a.Atoms < 0 || a.BasisPoints < 0 {
			return errors.New("invalid CARROT allocation")
		}
		if _, ok := seen[a.Account]; ok {
			return fmt.Errorf("duplicate allocation account %q", a.Account)
		}
		seen[a.Account] = struct{}{}
		atoms += a.Atoms
		bps += a.BasisPoints
	}
	if atoms != MaxSupplyAtoms || bps != 10_000 {
		return fmt.Errorf("allocation totals atoms=%d bps=%d", atoms, bps)
	}
	if allocationAtoms(p, AccountFounder) != FounderAllocationAtoms || allocationAtoms(p, AccountParticipation) != ParticipationReserveAtoms {
		return errors.New("founder or participation allocation drift")
	}
	if p.Fees.BurnBasisPoints != 0 || p.Fees.RecipientMode != "finality-signers-equal" || p.Fees.MinimumFeeAtoms < 0 {
		return errors.New("unexpected CARROT fee policy")
	}
	if p.Custody.FounderThreshold != 2 || p.Custody.FounderKeyCount != 3 || p.Custody.TreasuryThreshold != 3 || p.Custody.TreasuryKeyCount != 5 {
		return errors.New("unexpected CARROT custody threshold")
	}
	if p.Custody.TreasurySpendingEnabled {
		return errors.New("v0.10 treasury spending must remain disabled")
	}
	return nil
}

func (p Policy) Hash() string {
	raw, _ := json.Marshal(p)
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func allocationAtoms(p Policy, account string) int64 {
	for _, a := range p.Allocations {
		if a.Account == account {
			return a.Atoms
		}
	}
	return 0
}
