package publictestnet

import (
	"errors"
	"testing"
)

func TestMempoolQueuesContiguousNoncesAndReplacesByFee(t *testing.T) {
	sender, _ := GenerateWallet()
	to1, _ := GenerateWallet()
	to2, _ := GenerateWallet()
	to3, _ := GenerateWallet()
	pool := NewMempool("public", 64)

	tx0, _ := NewSignedTransaction("public", to1.Address(), 10, 0, 0, 100, sender)
	tx1, _ := NewSignedTransaction("public", to2.Address(), 10, 1, 1, 100, sender)
	tx2, _ := NewSignedTransaction("public", to3.Address(), 10, 0, 2, 100, sender)
	for _, tx := range []SignedTransaction{tx0, tx1, tx2} {
		if err := pool.Admit(tx, 1, 0); err != nil {
			t.Fatal(err)
		}
	}
	if pool.Count() != 3 {
		t.Fatalf("count=%d", pool.Count())
	}

	underpriced, _ := NewSignedTransaction("public", to3.Address(), 11, 1, 1, 100, sender)
	if err := pool.Admit(underpriced, 1, 0); !errors.Is(err, ErrReplacementUnderpriced) {
		t.Fatalf("underpriced replacement err=%v", err)
	}
	replacement, _ := NewSignedTransaction("public", to3.Address(), 11, 2, 1, 100, sender)
	if err := pool.Admit(replacement, 1, 0); err != nil {
		t.Fatal(err)
	}
	if pool.Count() != 3 {
		t.Fatalf("replacement changed count=%d", pool.Count())
	}

	ops := pool.OperationsFrom(10, func(string) uint64 { return 0 })
	if len(ops) != 3 {
		t.Fatalf("operations=%d", len(ops))
	}
	got := make([]SignedTransaction, 0, len(ops))
	for _, op := range ops {
		tx, err := ParseTransactionOperation(op)
		if err != nil {
			t.Fatal(err)
		}
		got = append(got, tx)
	}
	if got[0].Nonce != 0 || got[1].Nonce != 1 || got[2].Nonce != 2 || got[1].ID != replacement.ID {
		t.Fatalf("unexpected nonce chain: %+v", got)
	}
}

func TestMempoolNonceWindowAndGapAreBounded(t *testing.T) {
	sender, _ := GenerateWallet()
	to, _ := GenerateWallet()
	pool := NewMempool("public", 64)

	tooFar, _ := NewSignedTransaction("public", to.Address(), 1, 0, MaxQueuedTransactionsPerSender, 100, sender)
	if err := pool.Admit(tooFar, 1, 0); !errors.Is(err, ErrInvalidNonce) {
		t.Fatalf("far nonce err=%v", err)
	}
	tx0, _ := NewSignedTransaction("public", to.Address(), 1, 0, 0, 100, sender)
	tx2, _ := NewSignedTransaction("public", to.Address(), 1, 0, 2, 100, sender)
	if err := pool.Admit(tx0, 1, 0); err != nil {
		t.Fatal(err)
	}
	if err := pool.Admit(tx2, 1, 0); err != nil {
		t.Fatal(err)
	}
	ops := pool.OperationsFrom(10, func(string) uint64 { return 0 })
	if len(ops) != 1 {
		t.Fatalf("gap should expose only nonce 0, got %d ops", len(ops))
	}
	parsed, _ := ParseTransactionOperation(ops[0])
	if parsed.Nonce != 0 {
		t.Fatalf("nonce=%d", parsed.Nonce)
	}
}
