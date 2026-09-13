package carrot

import (
	"errors"
	"testing"
)

func TestPolicyFreezesSupplyAndAllocations(t *testing.T) {
	p := DefaultPolicy()
	if err := p.Validate(); err != nil {
		t.Fatal(err)
	}
	if FounderAllocationAtoms != 420_000_000_000_000 {
		t.Fatalf("founder=%d", FounderAllocationAtoms)
	}
	if ParticipationReserveAtoms != 1_260_000_000_000_000 {
		t.Fatalf("participation=%d", ParticipationReserveAtoms)
	}
	if TotalScheduledParticipation() != ParticipationReserveAtoms {
		t.Fatalf("scheduled=%d", TotalScheduledParticipation())
	}
	const wantHash = "19433c11cd2e19e973c9f15a5168b2d9f9fe676e368cffc26d8a2d077cd7f496"
	if p.Hash() != wantHash {
		t.Fatalf("policy hash=%s want %s", p.Hash(), wantHash)
	}
}

func TestEraBudgetsDeclineAndExhaustReserveExactly(t *testing.T) {
	var total int64
	prior := int64(^uint64(0) >> 1)
	for era := uint32(0); era < IssuanceEraCount; era++ {
		budget := EraBudget(era)
		if budget <= 0 {
			t.Fatalf("era %d budget=%d", era, budget)
		}
		if era > 0 && era < IssuanceEraCount-1 && budget > prior {
			t.Fatalf("era %d increased", era)
		}
		total += budget
		prior = budget
	}
	if total != ParticipationReserveAtoms {
		t.Fatalf("total=%d", total)
	}
	if EraBudget(IssuanceEraCount) != 0 {
		t.Fatal("post-schedule budget")
	}
}

func TestRewardAtHeightDeclinesAcrossEraBoundary(t *testing.T) {
	first := RewardAtHeight(1)
	secondEra := RewardAtHeight(IssuanceEraEpochs + 1)
	if first <= 0 || secondEra <= 0 || secondEra >= first {
		t.Fatalf("first=%d second=%d", first, secondEra)
	}
}

func TestFounderCliffAndLinearVesting(t *testing.T) {
	if FounderUnlockedAt(FounderCliffEpochs-1) != 0 {
		t.Fatal("unlocked before cliff")
	}
	atCliff := FounderUnlockedAt(FounderCliffEpochs)
	if atCliff != FounderAllocationAtoms/4 {
		t.Fatalf("cliff=%d", atCliff)
	}
	if FounderUnlockedAt(FounderVestingEpochs) != FounderAllocationAtoms {
		t.Fatal("not fully vested")
	}
}

func TestLedgerConservesSupplyAndDistributesRewardsDeterministically(t *testing.T) {
	l, err := NewLedger(DefaultPolicy())
	if err != nil {
		t.Fatal(err)
	}
	reward := RewardAtHeight(1)
	got, err := l.ReleaseParticipation(1, []string{"b", "a", "a", "c"})
	if err != nil {
		t.Fatal(err)
	}
	if got["a"]+got["b"]+got["c"] != reward {
		t.Fatalf("distribution=%v reward=%d", got, reward)
	}
	if got["a"] < got["c"] {
		t.Fatalf("remainder not deterministic: %v", got)
	}
	if err := l.ValidateConservation(); err != nil {
		t.Fatal(err)
	}
	report := l.SupplyReport()
	if report.ParticipationReleasedAtoms != reward {
		t.Fatalf("report=%+v", report)
	}
}

func TestParticipationReleaseCannotReplayHeight(t *testing.T) {
	l, _ := NewLedger(DefaultPolicy())
	if _, err := l.ReleaseParticipation(1, []string{"a"}); err != nil {
		t.Fatal(err)
	}
	if _, err := l.ReleaseParticipation(1, []string{"a"}); !errors.Is(err, ErrInvalidHeight) {
		t.Fatalf("err=%v", err)
	}
}

func TestFounderTransferRespectsVestingAndFeesDoNotChangeSupply(t *testing.T) {
	l, _ := NewLedger(DefaultPolicy())
	if err := l.Transfer(AccountFounder, "alice", AtomsPerCARROT, 0, nil); !errors.Is(err, ErrFounderLocked) {
		t.Fatalf("err=%v", err)
	}
	l.height = FounderCliffEpochs
	if err := l.Transfer(AccountFounder, "alice", 10*AtomsPerCARROT, 3, []string{"validator-b", "validator-a"}); err != nil {
		t.Fatal(err)
	}
	if l.Balance("validator-a") != 2 || l.Balance("validator-b") != 1 {
		t.Fatalf("fees a=%d b=%d", l.Balance("validator-a"), l.Balance("validator-b"))
	}
	if err := l.ValidateConservation(); err != nil {
		t.Fatal(err)
	}
}

func TestTreasuryGenericSpendDisabled(t *testing.T) {
	l, _ := NewLedger(DefaultPolicy())
	if err := l.Transfer(AccountCommunity, "alice", AtomsPerCARROT, 0, nil); !errors.Is(err, ErrReservedAccount) {
		t.Fatalf("err=%v", err)
	}
}
