package testnet

import "testing"

func TestProposalAndVoteEquivocationEvidence(t *testing.T) {
	f := makeFixture(t)
	engine, _ := NewEngine(f.genesis)
	committee := SelectCommittee(f.validators, f.genesis.Config.CommitteeTarget, engine.Status().FinalizedHash, 1, 0)
	proposer, _ := ExpectedProposer(committee, 1, 0)
	idx := validatorIndex(f.validators, proposer.ID)

	blockA, err := engine.Draft(0, proposer.ID, nil, f.now)
	if err != nil {
		t.Fatal(err)
	}
	proposalA, _ := BuildProposal(blockA, proposer, f.keys[idx])
	blockB, err := engine.Draft(0, proposer.ID, []Operation{{Type: "noop", Key: "other"}}, f.now)
	if err != nil {
		t.Fatal(err)
	}
	proposalB, _ := BuildProposal(blockB, proposer, f.keys[idx])
	proposalEvidence, err := BuildProposalEquivocationEvidence(proposalA, proposalB, committee)
	if err != nil {
		t.Fatal(err)
	}
	if err := VerifyEvidence(proposalEvidence, committee); err != nil {
		t.Fatal(err)
	}

	validator := committee[0]
	validatorIdx := validatorIndex(f.validators, validator.ID)
	voteA, _ := BuildVote(blockA, validator, f.keys[validatorIdx])
	voteB, _ := BuildVote(blockB, validator, f.keys[validatorIdx])
	voteEvidence, err := BuildVoteEquivocationEvidence(voteA, voteB, committee)
	if err != nil {
		t.Fatal(err)
	}
	if err := VerifyEvidence(voteEvidence, committee); err != nil {
		t.Fatal(err)
	}
}
