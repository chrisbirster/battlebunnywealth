package testnet

import (
	"fmt"
	"time"
)

func RunSmokeCluster(blocks int) error {
	if blocks <= 0 {
		blocks = 5
	}
	now := time.Unix(1_800_000_000, 0).UTC()
	validators := make([]Validator, 4)
	keys := make([]any, 4)
	for i := 0; i < 4; i++ {
		pub, priv, err := GenerateValidatorKey(AlgorithmEd25519)
		if err != nil {
			return err
		}
		validators[i] = Validator{ID: fmt.Sprintf("validator-%02d", i+1), AccountID: fmt.Sprintf("genesis-%d", i+1), DeviceID: fmt.Sprintf("phone-%d", i+1), Algorithm: AlgorithmEd25519, PublicKey: pub, Authority: 1000}
		keys[i] = priv
	}
	genesis, err := NewGenesis("bbw-pop-testnet-v1", now, DefaultProtocolConfig(), validators)
	if err != nil {
		return err
	}
	validators = genesis.Validators
	engines := make([]*Engine, 5)
	for i := range engines {
		engines[i], err = NewEngine(genesis)
		if err != nil {
			return err
		}
	}
	for h := 0; h < blocks; h++ {
		committee := SelectCommittee(validators, genesis.Config.CommitteeTarget, engines[0].Status().FinalizedHash, uint64(h+1), 0)
		proposer, _ := ExpectedProposer(committee, uint64(h+1), 0)
		idx := -1
		for i, v := range validators {
			if v.ID == proposer.ID {
				idx = i
				break
			}
		}
		block, err := engines[0].Draft(0, proposer.ID, nil, now.Add(time.Duration(h+1)*time.Minute))
		if err != nil {
			return err
		}
		proposal, err := BuildProposal(block, proposer, keys[idx])
		if err != nil {
			return err
		}
		for _, e := range engines {
			if err := e.HandleProposal(proposal); err != nil {
				return err
			}
		}
		for vi, v := range validators[:3] {
			vote, err := BuildVote(block, v, keys[vi])
			if err != nil {
				return err
			}
			for _, e := range engines {
				_, err := e.HandleVote(vote)
				if err != nil {
					return err
				}
			}
		}
		want := engines[0].Status()
		for i, e := range engines[1:] {
			got := e.Status()
			if got.Height != want.Height || got.FinalizedHash != want.FinalizedHash {
				return fmt.Errorf("node %d diverged at height %d", i+2, want.Height)
			}
		}
	}
	return nil
}
