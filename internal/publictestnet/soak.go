package publictestnet

import (
	"fmt"
	"sort"
	"time"

	"github.com/chrisbirster/battlebunnywealth/internal/carrot"
	"github.com/chrisbirster/battlebunnywealth/internal/testnet"
)

type SoakReport struct {
	Blocks                    int      `json:"blocks"`
	FinalHeight               uint64   `json:"finalHeight"`
	FinalStateRoot            string   `json:"finalStateRoot"`
	FinalizedHash             string   `json:"finalizedHash"`
	Restarts                  int      `json:"restarts"`
	TemporaryPartitions       int      `json:"temporaryPartitions"`
	AlternateQuorumRounds     int      `json:"alternateQuorumRounds"`
	ValidatorActivationHeight uint64   `json:"validatorActivationHeight"`
	UpgradeActivationHeight   uint64   `json:"upgradeActivationHeight"`
	FinalProtocolVersion      int      `json:"finalProtocolVersion"`
	FinalValidatorIDs         []string `json:"finalValidatorIds"`
}

// RunSoak executes a deterministic multi-node public-testnet workload. It
// intentionally crosses both v0.13 activation boundaries, isolates one full
// node periodically, varies the valid quorum subset, restarts/replays a node,
// and checks state-root convergence after every finalized block.
func RunSoak(blocks int) (SoakReport, error) {
	minimum := int(2 + MinUpgradeNoticeBlocks + 4)
	if blocks < minimum {
		blocks = minimum
	}
	now := time.Unix(1_800_000_000, 0).UTC()

	allValidators := make([]testnet.Validator, 5)
	keys := map[string]any{}
	for i := 0; i < len(allValidators); i++ {
		pub, priv, err := testnet.GenerateValidatorKey(testnet.AlgorithmEd25519)
		if err != nil {
			return SoakReport{}, err
		}
		wallet, err := GenerateWallet()
		if err != nil {
			return SoakReport{}, err
		}
		id := fmt.Sprintf("validator-%02d", i+1)
		allValidators[i] = testnet.Validator{
			ID:            id,
			AccountID:     fmt.Sprintf("account-%02d", i+1),
			DeviceID:      fmt.Sprintf("phone-%02d", i+1),
			Algorithm:     testnet.AlgorithmEd25519,
			PublicKey:     pub,
			RewardAddress: wallet.Address(),
			Authority:     1000,
			Active:        true,
		}
		keys[id] = priv
	}

	genesis, err := testnet.NewGenesis("bbw-public-soak-v1", now, testnet.DefaultProtocolConfig(), allValidators[:4])
	if err != nil {
		return SoakReport{}, err
	}

	const nodeCount = 5
	engines := make([]*testnet.Engine, nodeCount)
	states := make([]*ConsensusState, nodeCount)
	newNode := func() (*testnet.Engine, *ConsensusState, error) {
		state, err := NewConsensusState(genesis, carrot.DefaultPolicy())
		if err != nil {
			return nil, nil, err
		}
		engine, err := testnet.NewEngineWithStateMachine(genesis, state)
		if err != nil {
			return nil, nil, err
		}
		return engine, state, nil
	}
	for i := range engines {
		engine, state, err := newNode()
		if err != nil {
			return SoakReport{}, err
		}
		engines[i], states[i] = engine, state
	}

	validatorActivation := uint64(2 + MinValidatorSetNoticeBlocks)
	upgradeActivation := uint64(2 + MinUpgradeNoticeBlocks)
	newSet := []testnet.Validator{genesis.Validators[0], genesis.Validators[1], genesis.Validators[2], allValidators[4]}
	for i := range newSet {
		newSet[i].Active = true
	}
	setPlan := NewValidatorSetPlan(genesis.NetworkID, validatorActivation, newSet)
	setOp, err := ValidatorSetCommitmentOperation(setPlan)
	if err != nil {
		return SoakReport{}, err
	}
	upgradePlan := NewUpgradePlan(genesis.NetworkID, testnet.ProtocolVersion, testnet.ProtocolVersion+1, upgradeActivation, carrot.DefaultPolicy().Hash(), "v0.13.0")
	upgradeOp, err := UpgradeCommitmentOperation(upgradePlan)
	if err != nil {
		return SoakReport{}, err
	}

	report := SoakReport{
		Blocks:                    blocks,
		ValidatorActivationHeight: validatorActivation,
		UpgradeActivationHeight:   upgradeActivation,
	}
	var previous testnet.FinalizedBlock

	for height := uint64(1); height <= uint64(blocks); height++ {
		ops := []testnet.Operation(nil)
		if height > 1 {
			settlementOp, err := SettlementOperation(previous)
			if err != nil {
				return report, err
			}
			ops = append(ops, settlementOp)
		}
		if height == 2 {
			ops = append(ops, setOp, upgradeOp)
		}

		validators := engines[0].ValidatorsAtHeight(height)
		committee := testnet.SelectCommittee(validators, genesis.Config.CommitteeTarget, engines[0].Status().FinalizedHash, height, 0)
		proposer, ok := testnet.ExpectedProposer(committee, height, 0)
		if !ok {
			return report, fmt.Errorf("height %d: no proposer", height)
		}
		privateKey, ok := keys[proposer.ID]
		if !ok {
			return report, fmt.Errorf("height %d: missing proposer key %s", height, proposer.ID)
		}
		block, err := engines[0].Draft(0, proposer.ID, ops, now.Add(time.Duration(height)*time.Minute))
		if err != nil {
			return report, fmt.Errorf("height %d draft: %w", height, err)
		}
		proposal, err := testnet.BuildProposal(block, proposer, privateKey)
		if err != nil {
			return report, err
		}

		offline := -1
		if height%41 == 0 {
			offline = nodeCount - 1
			report.TemporaryPartitions++
		}
		for i, engine := range engines {
			if i == offline {
				continue
			}
			if err := engine.HandleProposal(proposal); err != nil {
				return report, fmt.Errorf("height %d node %d proposal: %w", height, i, err)
			}
		}

		quorum := testnet.QuorumFor(genesis.Config, len(committee))
		signers := append([]testnet.Validator(nil), committee[:quorum]...)
		if height%53 == 0 && len(committee) > quorum {
			signers = append([]testnet.Validator(nil), committee[len(committee)-quorum:]...)
			report.AlternateQuorumRounds++
		}
		for _, validator := range signers {
			key, ok := keys[validator.ID]
			if !ok {
				return report, fmt.Errorf("height %d: missing validator key %s", height, validator.ID)
			}
			vote, err := testnet.BuildVote(block, validator, key)
			if err != nil {
				return report, err
			}
			for i, engine := range engines {
				if i == offline {
					continue
				}
				if _, err := engine.HandleVote(vote); err != nil {
					return report, fmt.Errorf("height %d node %d vote: %w", height, i, err)
				}
			}
		}

		finalized := engines[0].Finalized()
		if len(finalized) != int(height) {
			return report, fmt.Errorf("height %d did not finalize", height)
		}
		previous = finalized[len(finalized)-1]
		if offline >= 0 {
			if err := engines[offline].ImportFinalized(previous); err != nil {
				return report, fmt.Errorf("height %d partition catch-up: %w", height, err)
			}
		}

		if height%97 == 0 {
			engine, state, err := newNode()
			if err != nil {
				return report, err
			}
			for _, finalizedBlock := range engines[0].Finalized() {
				if err := engine.ImportFinalized(finalizedBlock); err != nil {
					return report, fmt.Errorf("height %d restart replay at block %d: %w", height, finalizedBlock.Block.Height, err)
				}
			}
			engines[nodeCount-1], states[nodeCount-1] = engine, state
			report.Restarts++
		}

		want := engines[0].Status()
		for i := 1; i < len(engines); i++ {
			got := engines[i].Status()
			if got.Height != want.Height || got.FinalizedHash != want.FinalizedHash || got.StateRoot != want.StateRoot {
				return report, fmt.Errorf("height %d node %d diverged", height, i)
			}
		}
	}

	status := engines[0].Status()
	report.FinalHeight = status.Height
	report.FinalStateRoot = status.StateRoot
	report.FinalizedHash = status.FinalizedHash
	report.FinalProtocolVersion = engines[0].ProtocolVersionAtHeight(status.Height)
	for _, validator := range engines[0].ValidatorsAtHeight(status.Height) {
		report.FinalValidatorIDs = append(report.FinalValidatorIDs, validator.ID)
	}
	sort.Strings(report.FinalValidatorIDs)

	if report.FinalHeight != uint64(blocks) {
		return report, fmt.Errorf("final height=%d want=%d", report.FinalHeight, blocks)
	}
	if report.FinalProtocolVersion != testnet.ProtocolVersion+1 {
		return report, fmt.Errorf("protocol upgrade did not activate: version=%d", report.FinalProtocolVersion)
	}
	if hasValidatorID(engines[0].ValidatorsAtHeight(validatorActivation), genesis.Validators[3].ID) {
		return report, fmt.Errorf("removed genesis validator remained active")
	}
	if !hasValidatorID(engines[0].ValidatorsAtHeight(validatorActivation), allValidators[4].ID) {
		return report, fmt.Errorf("replacement validator did not activate")
	}
	for i, state := range states {
		if err := state.ValidateConservation(); err != nil {
			return report, fmt.Errorf("node %d supply: %w", i, err)
		}
	}
	return report, nil
}

func hasValidatorID(validators []testnet.Validator, id string) bool {
	for _, validator := range validators {
		if validator.ID == id {
			return true
		}
	}
	return false
}
