package testnet

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSyncFromWithStorePersistsImportedFinality(t *testing.T) {
	f := makeFixture(t)
	source, _ := NewEngine(f.genesis)
	block, proposal := f.proposal(t, source, 0)
	if err := source.HandleProposal(proposal); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		vote, _ := BuildVote(block, f.validators[i], f.keys[i])
		if _, err := source.HandleVote(vote); err != nil {
			t.Fatal(err)
		}
	}
	server := httptest.NewServer(NewHTTPNode(source, nil, nil))
	defer server.Close()

	dir := t.TempDir()
	store, err := OpenStore(dir, f.genesis)
	if err != nil {
		t.Fatal(err)
	}
	stale, _ := NewEngine(f.genesis)
	if err := SyncFromWithStore(server.URL, server.Client(), stale, store); err != nil {
		t.Fatal(err)
	}
	if stale.Height() != 1 {
		t.Fatalf("synced height=%d", stale.Height())
	}

	restarted, _ := NewEngine(f.genesis)
	if err := store.Restore(restarted); err != nil {
		t.Fatal(err)
	}
	if restarted.Height() != 1 || restarted.Status().FinalizedHash != source.Status().FinalizedHash {
		t.Fatalf("durable catch-up lost: restarted=%+v source=%+v", restarted.Status(), source.Status())
	}
}

func TestStoreIgnoresStaleRoundStateAfterFinality(t *testing.T) {
	f := makeFixture(t)
	engine, _ := NewEngine(f.genesis)
	block, proposal := f.proposal(t, engine, 0)
	if err := engine.HandleProposal(proposal); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(t.TempDir(), 0o700); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	store, err := OpenStore(dir, f.genesis)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.SaveRoundProgress(engine.RoundProgress()); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		vote, _ := BuildVote(block, f.validators[i], f.keys[i])
		if _, err := engine.HandleVote(vote); err != nil {
			t.Fatal(err)
		}
	}
	if err := store.Append(engine.Finalized()[0]); err != nil {
		t.Fatal(err)
	}

	restarted, _ := NewEngine(f.genesis)
	if err := store.Restore(restarted); err != nil {
		t.Fatal(err)
	}
	if restarted.Height() != 1 || restarted.CurrentRound() != 0 {
		t.Fatalf("status after stale round cleanup=%+v", restarted.Status())
	}
	if _, err := os.Stat(filepath.Join(dir, "round-state.json")); !os.IsNotExist(err) {
		t.Fatalf("stale round state still exists: err=%v", err)
	}
}
