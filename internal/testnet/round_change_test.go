package testnet

import (
	"errors"
	"testing"
	"time"
)

func validatorIndex(items []Validator, id string) int {
	for i, validator := range items {
		if validator.ID == id {
			return i
		}
	}
	return -1
}

func advanceRoundOne(t *testing.T, f fixture, engine *Engine, locks map[string]ValidatorLock) RoundCertificate {
	t.Helper()
	var cert RoundCertificate
	for i := 0; i < 3; i++ {
		validator := f.validators[i]
		lock := locks[validator.ID]
		var (
			change RoundChange
			err    error
		)
		if lock.ValueHash != "" {
			change, err = BuildRoundChange(f.genesis.NetworkID, engine.Height()+1, 0, validator, f.keys[i], lock.Round, lock.ValueHash, lock.Proof)
		} else {
			change, err = BuildRoundChange(f.genesis.NetworkID, engine.Height()+1, 0, validator, f.keys[i], 0, "")
		}
		if err != nil {
			t.Fatal(err)
		}
		advanced, got, err := engine.HandleRoundChange(change)
		if err != nil {
			t.Fatal(err)
		}
		if i < 2 && advanced {
			t.Fatal("round advanced before quorum")
		}
		if advanced {
			cert = got
		}
	}
	if engine.CurrentRound() != 1 || cert.Round != 1 {
		t.Fatalf("round=%d cert=%+v", engine.CurrentRound(), cert)
	}
	return cert
}

func TestRoundChangeAdvancesAndFinalizesNextRound(t *testing.T) {
	f := makeFixture(t)
	engine, _ := NewEngine(f.genesis)
	advanceRoundOne(t, f, engine, nil)

	committee := SelectCommittee(f.validators, f.genesis.Config.CommitteeTarget, engine.Status().FinalizedHash, 1, 0)
	proposer, ok := ExpectedProposer(committee, 1, 1)
	if !ok {
		t.Fatal("missing round-1 proposer")
	}
	idx := validatorIndex(f.validators, proposer.ID)
	block, err := engine.Draft(1, proposer.ID, nil, f.now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if block.RoundCertificate == nil || block.RoundCertificate.Round != 1 {
		t.Fatal("round-1 block missing timeout certificate")
	}
	proposal, err := BuildProposal(block, proposer, f.keys[idx])
	if err != nil {
		t.Fatal(err)
	}
	if err := engine.HandleProposal(proposal); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		vote, _ := BuildVote(block, f.validators[i], f.keys[i])
		finalized, err := engine.HandleVote(vote)
		if err != nil {
			t.Fatal(err)
		}
		if i == 2 && !finalized {
			t.Fatal("round-1 block did not finalize")
		}
	}
	if engine.Height() != 1 || engine.CurrentRound() != 0 {
		t.Fatalf("height=%d round=%d", engine.Height(), engine.CurrentRound())
	}
}

func TestRoundChangeCarriesQuorumIntersectingLockedValue(t *testing.T) {
	f := makeFixture(t)
	engine, _ := NewEngine(f.genesis)
	block0, proposal0 := f.proposal(t, engine, 0)
	if err := engine.HandleProposal(proposal0); err != nil {
		t.Fatal(err)
	}
	locks := map[string]ValidatorLock{}
	for i := 0; i < 2; i++ {
		vote, _ := BuildVote(block0, f.validators[i], f.keys[i])
		finalized, err := engine.HandleVote(vote)
		if err != nil || finalized {
			t.Fatalf("vote %d finalized=%v err=%v", i, finalized, err)
		}
		locks[f.validators[i].ID] = ValidatorLock{ValidatorID: f.validators[i].ID, Round: 0, ValueHash: vote.ValueHash, Proof: vote}
	}
	value := blockValueHash(block0)
	cert := advanceRoundOne(t, f, engine, locks)
	if cert.LockedValueHash != value {
		t.Fatalf("locked value=%s want=%s", cert.LockedValueHash, value)
	}

	committee := SelectCommittee(f.validators, f.genesis.Config.CommitteeTarget, engine.Status().FinalizedHash, 1, 0)
	proposer, _ := ExpectedProposer(committee, 1, 1)
	if _, err := engine.Draft(1, proposer.ID, []Operation{{Type: "noop", Key: "different"}}, f.now.Add(2*time.Minute)); !errors.Is(err, ErrLockedValue) {
		t.Fatalf("different value err=%v", err)
	}
	block1, err := engine.Draft(1, proposer.ID, nil, f.now.Add(2*time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if blockValueHash(block1) != value {
		t.Fatal("locked value changed across proposer rotation")
	}
	idx := validatorIndex(f.validators, proposer.ID)
	proposal1, _ := BuildProposal(block1, proposer, f.keys[idx])
	if err := engine.HandleProposal(proposal1); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		vote, _ := BuildVote(block1, f.validators[i], f.keys[i])
		if _, err := engine.HandleVote(vote); err != nil {
			t.Fatal(err)
		}
	}
	if engine.Height() != 1 {
		t.Fatal("locked value did not finalize in later round")
	}
}

func TestSingleProvenLockDoesNotControlRoundCertificate(t *testing.T) {
	f := makeFixture(t)
	engine, _ := NewEngine(f.genesis)
	block0, proposal0 := f.proposal(t, engine, 0)
	if err := engine.HandleProposal(proposal0); err != nil {
		t.Fatal(err)
	}
	vote, _ := BuildVote(block0, f.validators[0], f.keys[0])
	if finalized, err := engine.HandleVote(vote); err != nil || finalized {
		t.Fatalf("partial vote finalized=%v err=%v", finalized, err)
	}
	locks := map[string]ValidatorLock{
		f.validators[0].ID: {ValidatorID: f.validators[0].ID, Round: 0, ValueHash: vote.ValueHash, Proof: vote},
	}
	cert := advanceRoundOne(t, f, engine, locks)
	if cert.LockedValueHash != "" {
		t.Fatalf("single lock unexpectedly controlled certificate: %+v", cert)
	}

	committee := SelectCommittee(f.validators, f.genesis.Config.CommitteeTarget, engine.Status().FinalizedHash, 1, 0)
	proposer, _ := ExpectedProposer(committee, 1, 1)
	if _, err := engine.Draft(1, proposer.ID, []Operation{{Type: "noop", Key: "new-value"}}, f.now.Add(time.Minute)); err != nil {
		t.Fatalf("single lock stalled later-round proposal: %v", err)
	}
}

func TestRoundChangeRejectsUnprovedAndTamperedLocks(t *testing.T) {
	f := makeFixture(t)
	engine, _ := NewEngine(f.genesis)
	block, proposal := f.proposal(t, engine, 0)
	if err := engine.HandleProposal(proposal); err != nil {
		t.Fatal(err)
	}
	vote, _ := BuildVote(block, f.validators[0], f.keys[0])
	if _, err := engine.HandleVote(vote); err != nil {
		t.Fatal(err)
	}

	unproved := RoundChange{NetworkID: f.genesis.NetworkID, Height: 1, FromRound: 0, ToRound: 1, ValidatorID: f.validators[0].ID, LockedRound: 0, LockedValueHash: vote.ValueHash}
	unproved.Signature, _ = Sign(f.validators[0].Algorithm, f.keys[0], roundChangeSigningMessage(unproved))
	if _, _, err := engine.HandleRoundChange(unproved); !errors.Is(err, ErrInvalidRoundChange) {
		t.Fatalf("unproved lock err=%v", err)
	}

	tamperedProof := vote
	tamperedProof.ValueHash = hashText("forged")
	forged := RoundChange{NetworkID: f.genesis.NetworkID, Height: 1, FromRound: 0, ToRound: 1, ValidatorID: f.validators[0].ID, LockedRound: 0, LockedValueHash: tamperedProof.ValueHash, LockProof: &tamperedProof}
	forged.Signature, _ = Sign(f.validators[0].Algorithm, f.keys[0], roundChangeSigningMessage(forged))
	if _, _, err := engine.HandleRoundChange(forged); !errors.Is(err, ErrInvalidSignature) {
		t.Fatalf("tampered lock proof err=%v", err)
	}
}

func TestValidatorCannotOmitObservedLock(t *testing.T) {
	f := makeFixture(t)
	engine, _ := NewEngine(f.genesis)
	block, proposal := f.proposal(t, engine, 0)
	if err := engine.HandleProposal(proposal); err != nil {
		t.Fatal(err)
	}
	vote, _ := BuildVote(block, f.validators[0], f.keys[0])
	if _, err := engine.HandleVote(vote); err != nil {
		t.Fatal(err)
	}
	change, _ := BuildRoundChange(f.genesis.NetworkID, 1, 0, f.validators[0], f.keys[0], 0, "")
	if _, _, err := engine.HandleRoundChange(change); !errors.Is(err, ErrLockedValue) {
		t.Fatalf("lock omission err=%v", err)
	}
}

func TestRoundProgressSurvivesRestart(t *testing.T) {
	f := makeFixture(t)
	engine, _ := NewEngine(f.genesis)
	block, proposal := f.proposal(t, engine, 0)
	if err := engine.HandleProposal(proposal); err != nil {
		t.Fatal(err)
	}
	vote, _ := BuildVote(block, f.validators[0], f.keys[0])
	if _, err := engine.HandleVote(vote); err != nil {
		t.Fatal(err)
	}
	locks := map[string]ValidatorLock{
		f.validators[0].ID: {ValidatorID: f.validators[0].ID, Round: 0, ValueHash: vote.ValueHash, Proof: vote},
	}
	advanceRoundOne(t, f, engine, locks)

	store, err := OpenStore(t.TempDir(), f.genesis)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.SaveRoundProgress(engine.RoundProgress()); err != nil {
		t.Fatal(err)
	}
	restarted, _ := NewEngine(f.genesis)
	if err := store.Restore(restarted); err != nil {
		t.Fatal(err)
	}
	if restarted.CurrentRound() != 1 || restarted.Status().LockedValidators != 1 {
		t.Fatalf("restored status=%+v", restarted.Status())
	}
	committee := SelectCommittee(f.validators, f.genesis.Config.CommitteeTarget, restarted.Status().FinalizedHash, 1, 0)
	proposer, _ := ExpectedProposer(committee, 1, 1)
	if _, err := restarted.Draft(1, proposer.ID, nil, f.now.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
}

func TestProposerEquivocationRejected(t *testing.T) {
	f := makeFixture(t)
	engine, _ := NewEngine(f.genesis)
	committee := SelectCommittee(f.validators, f.genesis.Config.CommitteeTarget, engine.Status().FinalizedHash, 1, 0)
	proposer, _ := ExpectedProposer(committee, 1, 0)
	idx := validatorIndex(f.validators, proposer.ID)

	first, _ := engine.Draft(0, proposer.ID, nil, f.now)
	firstProposal, _ := BuildProposal(first, proposer, f.keys[idx])
	if err := engine.HandleProposal(firstProposal); err != nil {
		t.Fatal(err)
	}
	second, _ := engine.Draft(0, proposer.ID, []Operation{{Type: "noop", Key: "equivocation"}}, f.now.Add(time.Second))
	secondProposal, _ := BuildProposal(second, proposer, f.keys[idx])
	if err := engine.HandleProposal(secondProposal); !errors.Is(err, ErrEquivocation) {
		t.Fatalf("equivocation err=%v", err)
	}
}
