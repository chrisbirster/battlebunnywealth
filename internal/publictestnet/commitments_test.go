package publictestnet

import (
	"errors"
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
	status := recoveredState.TransitionStatus()
	if status.PendingValidatorSet == nil || status.PendingUpgrade == nil {
		t.Fatal("recovery lost pending transition commitments")
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

func finalizeWithRules(t *testing.T, f *executionFixture, keys map[string]any, operations []testnet.Operation) testnet.FinalizedBlock {
	t.Helper()
	height := f.engines[0].Height() + 1
	validators := f.engines[0].ValidatorsAtHeight(height)
	committee := testnet.SelectCommittee(validators, f.genesis.Config.CommitteeTarget, f.engines[0].Status().FinalizedHash, height, 0)
	proposer, ok := testnet.ExpectedProposer(committee, height, 0)
	if !ok {
		t.Fatal("no proposer")
	}
	key, ok := keys[proposer.ID]
	if !ok {
		t.Fatalf("missing proposer key %s", proposer.ID)
	}
	block, err := f.engines[0].Draft(0, proposer.ID, operations, f.now.AddDate(0, 0, int(height)))
	if err != nil {
		t.Fatal(err)
	}
	proposal, err := testnet.BuildProposal(block, proposer, key)
	if err != nil {
		t.Fatal(err)
	}
	for _, engine := range f.engines {
		if err := engine.HandleProposal(proposal); err != nil {
			t.Fatal(err)
		}
	}
	quorum := testnet.QuorumFor(f.genesis.Config, len(committee))
	for _, validator := range committee[:quorum] {
		privateKey, ok := keys[validator.ID]
		if !ok {
			t.Fatalf("missing validator key %s", validator.ID)
		}
		vote, err := testnet.BuildVote(block, validator, privateKey)
		if err != nil {
			t.Fatal(err)
		}
		for _, engine := range f.engines {
			if _, err := engine.HandleVote(vote); err != nil {
				t.Fatal(err)
			}
		}
	}
	blocks := f.engines[0].Finalized()
	return blocks[len(blocks)-1]
}

func TestCommittedValidatorAndProtocolTransitionsActivateAndReplay(t *testing.T) {
	f := newExecutionFixture(t, 2)
	keys := map[string]any{}
	for id, key := range f.keys {
		keys[id] = key
	}

	pub, priv, err := testnet.GenerateValidatorKey(testnet.AlgorithmEd25519)
	if err != nil {
		t.Fatal(err)
	}
	reward, err := GenerateWallet()
	if err != nil {
		t.Fatal(err)
	}
	replacement := testnet.Validator{
		ID:            "validator-replacement",
		AccountID:     "replacement-account",
		DeviceID:      "replacement-phone",
		Algorithm:     testnet.AlgorithmEd25519,
		PublicKey:     pub,
		RewardAddress: reward.Address(),
		Authority:     1000,
		Active:        true,
	}
	keys[replacement.ID] = priv

	newSet := append([]testnet.Validator(nil), f.validators[:3]...)
	for i := range newSet {
		newSet[i].Active = true
	}
	newSet = append(newSet, replacement)
	setActivation := uint64(2 + MinValidatorSetNoticeBlocks)
	upgradeActivation := uint64(2 + MinUpgradeNoticeBlocks)
	setPlan := NewValidatorSetPlan(f.genesis.NetworkID, setActivation, newSet)
	setOp, err := ValidatorSetCommitmentOperation(setPlan)
	if err != nil {
		t.Fatal(err)
	}
	upgradePlan := NewUpgradePlan(f.genesis.NetworkID, testnet.ProtocolVersion, testnet.ProtocolVersion+1, upgradeActivation, carrot.DefaultPolicy().Hash(), "v0.13.0")
	upgradeOp, err := UpgradeCommitmentOperation(upgradePlan)
	if err != nil {
		t.Fatal(err)
	}

	previous := finalizeWithRules(t, f, keys, nil)
	previous = finalizeWithRules(t, f, keys, []testnet.Operation{settlement(t, previous), setOp, upgradeOp})
	var blockBeforeSet, blockAtSet, blockAtUpgrade testnet.FinalizedBlock
	for f.engines[0].Height() < upgradeActivation+1 {
		height := f.engines[0].Height() + 1
		previous = finalizeWithRules(t, f, keys, []testnet.Operation{settlement(t, previous)})
		switch height {
		case setActivation - 1:
			blockBeforeSet = previous
		case setActivation:
			blockAtSet = previous
		case upgradeActivation:
			blockAtUpgrade = previous
		}
	}

	if blockBeforeSet.Block.Height != setActivation-1 || blockAtSet.Block.Height != setActivation {
		t.Fatal("validator activation boundary was not exercised")
	}
	oldID := f.validators[3].ID
	if containsValidator(f.engines[0].ValidatorsAtHeight(setActivation), oldID) {
		t.Fatal("removed validator remained active at activation height")
	}
	if !containsValidator(f.engines[0].ValidatorsAtHeight(setActivation), replacement.ID) {
		t.Fatal("replacement validator did not activate")
	}
	if !containsValidator(f.engines[0].ValidatorsAtHeight(setActivation-1), oldID) {
		t.Fatal("historical validator set was not retained")
	}
	if blockAtUpgrade.Block.Version != testnet.ProtocolVersion+1 {
		t.Fatalf("upgrade block version=%d", blockAtUpgrade.Block.Version)
	}
	if f.engines[0].ProtocolVersionAtHeight(upgradeActivation-1) != testnet.ProtocolVersion || f.engines[0].ProtocolVersionAtHeight(upgradeActivation) != testnet.ProtocolVersion+1 {
		t.Fatal("protocol version history is incorrect")
	}

	status := f.states[0].TransitionStatus()
	if status.PendingValidatorSet != nil || status.PendingUpgrade != nil || status.ActiveProtocolVersion != testnet.ProtocolVersion+1 || status.ActiveValidatorCount != 4 {
		t.Fatalf("unexpected transition status: %+v", status)
	}

	recoveredState, err := NewConsensusState(f.genesis, carrot.DefaultPolicy())
	if err != nil {
		t.Fatal(err)
	}
	recovered, err := testnet.NewEngineWithStateMachine(f.genesis, recoveredState)
	if err != nil {
		t.Fatal(err)
	}
	for _, block := range f.engines[0].Finalized() {
		if err := recovered.ImportFinalized(block); err != nil {
			t.Fatalf("replay height %d: %v", block.Block.Height, err)
		}
	}
	if recovered.Status().StateRoot != f.engines[0].Status().StateRoot || recovered.Status().FinalizedHash != f.engines[0].Status().FinalizedHash {
		t.Fatal("recovered node diverged after transition replay")
	}
	if !containsValidator(recovered.ValidatorsAtHeight(setActivation-1), oldID) || containsValidator(recovered.ValidatorsAtHeight(setActivation), oldID) {
		t.Fatal("recovered node lost historical validator-set boundary")
	}
	if recovered.ProtocolVersionAtHeight(upgradeActivation) != testnet.ProtocolVersion+1 {
		t.Fatal("recovered node lost protocol-upgrade activation")
	}

	nextHeight := f.engines[0].Height() + 1
	validators := f.engines[0].ValidatorsAtHeight(nextHeight)
	committee := testnet.SelectCommittee(validators, f.genesis.Config.CommitteeTarget, f.engines[0].Status().FinalizedHash, nextHeight, 0)
	proposer, _ := testnet.ExpectedProposer(committee, nextHeight, 0)
	block, err := f.engines[0].Draft(0, proposer.ID, []testnet.Operation{settlement(t, previous)}, f.now)
	if err != nil {
		t.Fatal(err)
	}
	proposal, _ := testnet.BuildProposal(block, proposer, keys[proposer.ID])
	if err := f.engines[0].HandleProposal(proposal); err != nil {
		t.Fatal(err)
	}
	oldVote, _ := testnet.BuildVote(block, f.validators[3], keys[oldID])
	if _, err := f.engines[0].HandleVote(oldVote); !errors.Is(err, testnet.ErrUnknownValidator) {
		t.Fatalf("removed validator vote err=%v", err)
	}
}

func containsValidator(validators []testnet.Validator, id string) bool {
	for _, validator := range validators {
		if validator.ID == id {
			return true
		}
	}
	return false
}
