package publictestnet

import (
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/chrisbirster/battlebunnywealth/internal/carrot"
	"github.com/chrisbirster/battlebunnywealth/internal/testnet"
)

type RoundChaosReport struct {
	Blocks                int    `json:"blocks"`
	Seed                  int64  `json:"seed"`
	FinalHeight           uint64 `json:"finalHeight"`
	FinalizedHash         string `json:"finalizedHash"`
	FinalStateRoot        string `json:"finalStateRoot"`
	RoundChanges          int    `json:"roundChanges"`
	LockedRoundChanges    int    `json:"lockedRoundChanges"`
	Partitions            int    `json:"partitions"`
	Restarts              int    `json:"restarts"`
	DuplicateMessages     int    `json:"duplicateMessages"`
	ReorderedDeliveries   int    `json:"reorderedDeliveries"`
	ClockSkewSamples      int    `json:"clockSkewSamples"`
}

type chaosNode struct {
	engine *testnet.Engine
	state  *ConsensusState
}

func RunRoundChaos(blocks int, seed int64) (RoundChaosReport, error) {
	if blocks < 120 {
		blocks = 120
	}
	rng := rand.New(rand.NewSource(seed))
	now := time.Unix(1_800_000_000, 0).UTC()
	validators := make([]testnet.Validator, 4)
	keys := map[string]any{}
	for i := 0; i < 4; i++ {
		pub, priv, err := testnet.GenerateValidatorKey(testnet.AlgorithmEd25519)
		if err != nil {
			return RoundChaosReport{}, err
		}
		id := fmt.Sprintf("validator-%02d", i+1)
		validators[i] = testnet.Validator{ID: id, AccountID: fmt.Sprintf("account-%02d", i+1), DeviceID: fmt.Sprintf("phone-%02d", i+1), Algorithm: testnet.AlgorithmEd25519, PublicKey: pub, Authority: 1000, Active: true}
		keys[id] = priv
	}
	genesis, err := testnet.NewGenesis("bbw-round-chaos-v1", now, testnet.DefaultProtocolConfig(), validators)
	if err != nil {
		return RoundChaosReport{}, err
	}

	newNode := func() (chaosNode, error) {
		state, err := NewConsensusState(genesis, carrot.DefaultPolicy())
		if err != nil {
			return chaosNode{}, err
		}
		engine, err := testnet.NewEngineWithStateMachine(genesis, state)
		if err != nil {
			return chaosNode{}, err
		}
		return chaosNode{engine: engine, state: state}, nil
	}

	nodes := make([]chaosNode, 5)
	for i := range nodes {
		nodes[i], err = newNode()
		if err != nil {
			return RoundChaosReport{}, err
		}
	}

	report := RoundChaosReport{Blocks: blocks, Seed: seed}
	var previous testnet.FinalizedBlock

	for height := uint64(1); height <= uint64(blocks); height++ {
		ops := []testnet.Operation(nil)
		if height > 1 {
			settlement, err := SettlementOperation(previous)
			if err != nil {
				return report, err
			}
			ops = append(ops, settlement)
		}

		offline := -1
		if height%17 == 0 {
			offline = len(nodes) - 1
			report.Partitions++
		}
		activeNodeIndexes := make([]int, 0, len(nodes))
		for i := range nodes {
			if i != offline {
				activeNodeIndexes = append(activeNodeIndexes, i)
			}
		}

		committee := testnet.SelectCommittee(nodes[0].engine.ValidatorsAtHeight(height), genesis.Config.CommitteeTarget, nodes[0].engine.Status().FinalizedHash, height, 0)
		quorum := testnet.QuorumFor(genesis.Config, len(committee))
		forceRoundChange := rng.Intn(3) == 0
		var block testnet.Block
		var proposal testnet.Proposal

		if forceRoundChange {
			proposer0, ok := testnet.ExpectedProposer(committee, height, 0)
			if !ok {
				return report, fmt.Errorf("height %d missing round-0 proposer", height)
			}
			skew := time.Duration(rng.Intn(121)-60) * time.Second
			block0, err := nodes[0].engine.Draft(0, proposer0.ID, ops, now.Add(time.Duration(height)*time.Minute+skew))
			if err != nil {
				return report, fmt.Errorf("height %d round-0 draft: %w", height, err)
			}
			report.ClockSkewSamples++
			proposal0, err := testnet.BuildProposal(block0, proposer0, keys[proposer0.ID])
			if err != nil {
				return report, err
			}
			delivery := append([]int(nil), activeNodeIndexes...)
			rng.Shuffle(len(delivery), func(i, j int) { delivery[i], delivery[j] = delivery[j], delivery[i] })
			report.ReorderedDeliveries++
			for _, index := range delivery {
				if err := nodes[index].engine.HandleProposal(proposal0); err != nil {
					return report, fmt.Errorf("height %d node %d round-0 proposal: %w", height, index, err)
				}
			}

			lockCount := rng.Intn(quorum)
			locked := map[string]testnet.ValidatorLock{}
			for i := 0; i < lockCount; i++ {
				validator := committee[i]
				vote, err := testnet.BuildVote(block0, validator, keys[validator.ID])
				if err != nil {
					return report, err
				}
				for _, index := range activeNodeIndexes {
					finalized, err := nodes[index].engine.HandleVote(vote)
					if err != nil || finalized {
						return report, fmt.Errorf("height %d partial lock vote node %d finalized=%v err=%v", height, index, finalized, err)
					}
				}
				locked[validator.ID] = testnet.ValidatorLock{ValidatorID: validator.ID, Round: 0, ValueHash: vote.ValueHash, Proof: vote}
			}
			if lockCount > 0 {
				report.LockedRoundChanges++
			}

			changes := make([]testnet.RoundChange, 0, quorum)
			for i := 0; i < quorum; i++ {
				validator := committee[i]
				lock := locked[validator.ID]
				var change testnet.RoundChange
				if lock.ValueHash != "" {
					change, err = testnet.BuildRoundChange(genesis.NetworkID, height, 0, validator, keys[validator.ID], lock.Round, lock.ValueHash, lock.Proof)
				} else {
					change, err = testnet.BuildRoundChange(genesis.NetworkID, height, 0, validator, keys[validator.ID], 0, "")
				}
				if err != nil {
					return report, err
				}
				changes = append(changes, change)
			}
			rng.Shuffle(len(changes), func(i, j int) { changes[i], changes[j] = changes[j], changes[i] })
			for _, index := range activeNodeIndexes {
				for _, change := range changes {
					_, _, err := nodes[index].engine.HandleRoundChange(change)
					if err != nil {
						return report, fmt.Errorf("height %d node %d round change: %w", height, index, err)
					}
				}
			}
			report.RoundChanges++

			proposer1, ok := testnet.ExpectedProposer(committee, height, 1)
			if !ok {
				return report, fmt.Errorf("height %d missing round-1 proposer", height)
			}
			block, err = nodes[0].engine.Draft(1, proposer1.ID, ops, now.Add(time.Duration(height)*time.Minute+time.Second))
			if err != nil {
				return report, fmt.Errorf("height %d round-1 draft: %w", height, err)
			}
			proposal, err = testnet.BuildProposal(block, proposer1, keys[proposer1.ID])
			if err != nil {
				return report, err
			}
		} else {
			proposer, ok := testnet.ExpectedProposer(committee, height, 0)
			if !ok {
				return report, fmt.Errorf("height %d missing proposer", height)
			}
			skew := time.Duration(rng.Intn(121)-60) * time.Second
			block, err = nodes[0].engine.Draft(0, proposer.ID, ops, now.Add(time.Duration(height)*time.Minute+skew))
			if err != nil {
				return report, err
			}
			report.ClockSkewSamples++
			proposal, err = testnet.BuildProposal(block, proposer, keys[proposer.ID])
			if err != nil {
				return report, err
			}
		}

		delivery := append([]int(nil), activeNodeIndexes...)
		rng.Shuffle(len(delivery), func(i, j int) { delivery[i], delivery[j] = delivery[j], delivery[i] })
		report.ReorderedDeliveries++
		for _, index := range delivery {
			if err := nodes[index].engine.HandleProposal(proposal); err != nil {
				return report, fmt.Errorf("height %d node %d final proposal: %w", height, index, err)
			}
		}

		signers := append([]testnet.Validator(nil), committee[:quorum]...)
		rng.Shuffle(len(signers), func(i, j int) { signers[i], signers[j] = signers[j], signers[i] })
		for voteIndex, validator := range signers {
			vote, err := testnet.BuildVote(block, validator, keys[validator.ID])
			if err != nil {
				return report, err
			}
			voteDelivery := append([]int(nil), activeNodeIndexes...)
			rng.Shuffle(len(voteDelivery), func(i, j int) { voteDelivery[i], voteDelivery[j] = voteDelivery[j], voteDelivery[i] })
			for _, index := range voteDelivery {
				_, err := nodes[index].engine.HandleVote(vote)
				if err != nil {
					return report, fmt.Errorf("height %d node %d vote: %w", height, index, err)
				}
			}
			if voteIndex == 0 && len(voteDelivery) > 0 {
				_, err := nodes[voteDelivery[0]].engine.HandleVote(vote)
				if !errors.Is(err, testnet.ErrDuplicateVote) {
					return report, fmt.Errorf("height %d duplicate vote err=%v", height, err)
				}
				report.DuplicateMessages++
			}
		}

		finalized := nodes[0].engine.Finalized()
		if len(finalized) != int(height) {
			return report, fmt.Errorf("height %d did not finalize", height)
		}
		previous = finalized[len(finalized)-1]
		if offline >= 0 {
			if err := nodes[offline].engine.ImportFinalized(previous); err != nil {
				return report, fmt.Errorf("height %d partition catch-up: %w", height, err)
			}
		}

		if height%29 == 0 {
			fresh, err := newNode()
			if err != nil {
				return report, err
			}
			for _, finalizedBlock := range nodes[0].engine.Finalized() {
				if err := fresh.engine.ImportFinalized(finalizedBlock); err != nil {
					return report, fmt.Errorf("height %d restart replay block %d: %w", height, finalizedBlock.Block.Height, err)
				}
			}
			nodes[len(nodes)-1] = fresh
			report.Restarts++
		}

		want := nodes[0].engine.Status()
		for i := 1; i < len(nodes); i++ {
			got := nodes[i].engine.Status()
			if got.Height != want.Height || got.FinalizedHash != want.FinalizedHash || got.StateRoot != want.StateRoot {
				return report, fmt.Errorf("height %d node %d diverged: got=%+v want=%+v", height, i, got, want)
			}
			if err := nodes[i].state.ValidateConservation(); err != nil {
				return report, fmt.Errorf("height %d node %d supply: %w", height, i, err)
			}
		}
	}

	status := nodes[0].engine.Status()
	report.FinalHeight = status.Height
	report.FinalizedHash = status.FinalizedHash
	report.FinalStateRoot = status.StateRoot
	if report.FinalHeight != uint64(blocks) || report.RoundChanges == 0 || report.Partitions == 0 || report.Restarts == 0 {
		return report, fmt.Errorf("chaos coverage incomplete: %+v", report)
	}
	return report, nil
}

func sortedValidatorIDs(items []testnet.Validator) []string {
	ids := make([]string, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ID)
	}
	sort.Strings(ids)
	return ids
}
