package publictestnet

import (
	"testing"

	"github.com/chrisbirster/battlebunnywealth/internal/carrot"
	"github.com/chrisbirster/battlebunnywealth/internal/testnet"
)

func TestFinalizedTransitionCommitmentsSurviveRecovery(t *testing.T) {
	f := newExecutionFixture(t, 2)
	b1, _ := f.finalize(t, nil)

	validators := append([]testnet.Validator(nil), f.validators...)
	for i := range validators {
		validators[i].Active = true
	}
	setPlan := NewValidatorSetPlan(f.genesis.NetworkID, 2+MinValidatorSetNoticeBlocks, validators)
	if err := setPlan.Validate(f.genesis.NetworkID, 2); err != nil {
		t.Fatal(err)
	}
	setOp, err := ValidatorSetCommitmentOperation(setPlan)
	if err != nil {
		t.Fatal(err)
	}

	upgradePlan := NewUpgradePlan(
		f.genesis.NetworkID,
		testnet.ProtocolVersion,
		testnet.ProtocolVersion+1,
		2+MinUpgradeNoticeBlocks,
		carrot.DefaultPolicy().Hash(),
		"v0.13.0",
	)
	if err := upgradePlan.Validate(f.genesis.NetworkID, testnet.ProtocolVersion, 2); err != nil {
		t.Fatal(err)
	}
	upgradeOp, err := UpgradeCommitmentOperation(upgradePlan)
	if err != nil {
		t.Fatal(err)
	}

	b2, _ := f.finalize(t, []testnet.Operation{settlement(t, b1), setOp, upgradeOp})
	if len(b2.Block.Operations) != 3 {
		t.Fatalf("operations=%d", len(b2.Block.Operations))
	}
	if _, err := ParseValidatorSetCommitmentOperation(b2.Block.Operations[1]); err != nil {
		t.Fatal(err)
	}
	if _, err := ParseUpgradeCommitmentOperation(b2.Block.Operations[2]); err != nil {
		t.Fatal(err)
	}
	if f.engines[0].Status().FinalizedHash != f.engines[1].Status().FinalizedHash {
		t.Fatal("nodes disagreed on finalized commitment block")
	}

	store, err := testnet.OpenStore(t.TempDir(), f.genesis)
	if err != nil {
		t.Fatal(err)
	}
	for _, block := range f.engines[0].Finalized() {
		if err := store.Append(block); err != nil {
			t.Fatal(err)
		}
	}
	recoveredState, err := NewConsensusState(f.genesis, carrot.DefaultPolicy())
	if err != nil {
		t.Fatal(err)
	}
	recovered, err := testnet.NewEngineWithStateMachine(f.genesis, recoveredState)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Restore(recovered); err != nil {
		t.Fatal(err)
	}
	if recovered.Status().FinalizedHash != b2.Block.Hash || recovered.Status().StateRoot != f.engines[0].Status().StateRoot {
		t.Fatal("recovery did not preserve finalized transition commitment chain")
	}
}

func TestTooSoonTransitionCommitmentsAreRejected(t *testing.T) {
	f := newExecutionFixture(t, 1)
	b1, _ := f.finalize(t, nil)
	p := f.expectedProposer(t)

	validators := append([]testnet.Validator(nil), f.validators...)
	for i := range validators {
		validators[i].Active = true
	}
	setPlan := NewValidatorSetPlan(f.genesis.NetworkID, 3, validators)
	setOp, _ := ValidatorSetCommitmentOperation(setPlan)
	if _, err := f.engines[0].Draft(0, p.ID, []testnet.Operation{settlement(t, b1), setOp}, f.now); err == nil {
		t.Fatal("too-soon validator-set commitment was draftable")
	}

	upgradePlan := NewUpgradePlan(f.genesis.NetworkID, testnet.ProtocolVersion, testnet.ProtocolVersion+1, 3, carrot.DefaultPolicy().Hash(), "v0.13.0")
	upgradeOp, _ := UpgradeCommitmentOperation(upgradePlan)
	if _, err := f.engines[0].Draft(0, p.ID, []testnet.Operation{settlement(t, b1), upgradeOp}, f.now); err == nil {
		t.Fatal("too-soon protocol-upgrade commitment was draftable")
	}
}
