package publictestnet

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"sort"

	"github.com/chrisbirster/battlebunnywealth/internal/testnet"
)

const (
	ValidatorSetCommitmentOperationType = "validator-set-commitment"
	UpgradeCommitmentOperationType      = "protocol-upgrade-commitment"
	MinValidatorSetNoticeBlocks  uint64 = 144
)

var ErrInvalidValidatorSetCommitment = errors.New("invalid validator-set commitment")

type ValidatorSetPlan struct {
	NetworkID        string              `json:"networkId"`
	ActivationHeight uint64              `json:"activationHeight"`
	Validators       []testnet.Validator `json:"validators"`
	Hash             string              `json:"hash"`
}

func NewValidatorSetPlan(networkID string, activationHeight uint64, validators []testnet.Validator) ValidatorSetPlan {
	normalized := append([]testnet.Validator(nil), validators...)
	sort.Slice(normalized, func(i, j int) bool { return normalized[i].ID < normalized[j].ID })
	p := ValidatorSetPlan{NetworkID: networkID, ActivationHeight: activationHeight, Validators: normalized}
	p.Hash = validatorSetPlanHash(p)
	return p
}

func (p ValidatorSetPlan) Validate(networkID string, currentHeight uint64) error {
	if p.NetworkID != networkID || p.ActivationHeight < currentHeight+MinValidatorSetNoticeBlocks || len(p.Validators) == 0 || p.Hash != validatorSetPlanHash(p) {
		return ErrInvalidValidatorSetCommitment
	}
	seenIDs := map[string]struct{}{}
	seenKeys := map[string]struct{}{}
	last := ""
	for _, v := range p.Validators {
		if v.ID == "" || v.Algorithm == "" || v.PublicKey == "" || v.Authority <= 0 || !v.Active || v.ID < last {
			return ErrInvalidValidatorSetCommitment
		}
		if v.RewardAddress != "" && !ValidAddress(v.RewardAddress) {
			return ErrInvalidValidatorSetCommitment
		}
		if _, ok := seenIDs[v.ID]; ok {
			return ErrInvalidValidatorSetCommitment
		}
		if _, ok := seenKeys[v.PublicKey]; ok {
			return ErrInvalidValidatorSetCommitment
		}
		seenIDs[v.ID] = struct{}{}
		seenKeys[v.PublicKey] = struct{}{}
		last = v.ID
	}
	return nil
}

func validatorSetPlanHash(p ValidatorSetPlan) string {
	p.Hash = ""
	raw, _ := json.Marshal(p)
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func ValidatorSetCommitmentOperation(plan ValidatorSetPlan) (testnet.Operation, error) {
	if plan.Hash == "" {
		return testnet.Operation{}, ErrInvalidValidatorSetCommitment
	}
	raw, err := json.Marshal(plan)
	if err != nil {
		return testnet.Operation{}, err
	}
	return testnet.Operation{Type: ValidatorSetCommitmentOperationType, Key: plan.Hash, Value: string(raw)}, nil
}

func ParseValidatorSetCommitmentOperation(op testnet.Operation) (ValidatorSetPlan, error) {
	if op.Type != ValidatorSetCommitmentOperationType || op.Key == "" || op.Value == "" {
		return ValidatorSetPlan{}, ErrInvalidValidatorSetCommitment
	}
	var plan ValidatorSetPlan
	if err := json.Unmarshal([]byte(op.Value), &plan); err != nil {
		return ValidatorSetPlan{}, err
	}
	if plan.Hash != op.Key || validatorSetPlanHash(plan) != plan.Hash {
		return ValidatorSetPlan{}, ErrInvalidValidatorSetCommitment
	}
	return plan, nil
}

func UpgradeCommitmentOperation(plan UpgradePlan) (testnet.Operation, error) {
	if plan.Hash == "" {
		return testnet.Operation{}, ErrInvalidUpgrade
	}
	raw, err := json.Marshal(plan)
	if err != nil {
		return testnet.Operation{}, err
	}
	return testnet.Operation{Type: UpgradeCommitmentOperationType, Key: plan.Hash, Value: string(raw)}, nil
}

func ParseUpgradeCommitmentOperation(op testnet.Operation) (UpgradePlan, error) {
	if op.Type != UpgradeCommitmentOperationType || op.Key == "" || op.Value == "" {
		return UpgradePlan{}, ErrInvalidUpgrade
	}
	var plan UpgradePlan
	if err := json.Unmarshal([]byte(op.Value), &plan); err != nil {
		return UpgradePlan{}, err
	}
	if plan.Hash != op.Key || upgradeHash(plan) != plan.Hash {
		return UpgradePlan{}, ErrInvalidUpgrade
	}
	return plan, nil
}
