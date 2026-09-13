package popsim

import "testing"

func TestBootstrapFarmImmediateVersusControlled(t *testing.T) {
	immediate := DefaultBootstrapAttackPolicy(100)
	immediate.ImmediateActivation = true
	a, err := SimulateBootstrapAttack(immediate)
	if err != nil {
		t.Fatal(err)
	}
	if a.FirstBlockingDay == nil || *a.FirstBlockingDay != 0 {
		t.Fatalf("immediate blocking=%v", a.FirstBlockingDay)
	}
	controlled := DefaultBootstrapAttackPolicy(100)
	b, err := SimulateBootstrapAttack(controlled)
	if err != nil {
		t.Fatal(err)
	}
	if b.FirstBlockingDay == nil || *b.FirstBlockingDay < 30 {
		t.Fatalf("controlled blocking=%v", b.FirstBlockingDay)
	}
	if *b.FirstBlockingDay <= *a.FirstBlockingDay {
		t.Fatal("activation policy did not delay farm")
	}
}
func TestBootstrapGate(t *testing.T) {
	if err := BootstrapGate(100); err != nil {
		t.Fatal(err)
	}
}
