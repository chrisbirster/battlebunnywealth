package publictestnet

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCheckpointVerifiesByIndependentReplay(t *testing.T) {
	f := newExecutionFixture(t, 1)
	b1, _ := f.finalize(t, nil)
	f.finalize(t, []testnet.Operation{settlement(t, b1)})

	checkpoint, err := BuildCheckpoint(f.engines[0], f.states[0])
	if err != nil {
		t.Fatal(err)
	}
	if checkpoint.Height != 2 || checkpoint.Hash == "" {
		t.Fatalf("checkpoint=%+v", checkpoint)
	}
	if err := VerifyCheckpoint(f.genesis, f.engines[0].Finalized(), checkpoint); err != nil {
		t.Fatal(err)
	}

	tampered := checkpoint
	tampered.StateRoot = "deadbeef"
	if err := VerifyCheckpoint(f.genesis, f.engines[0].Finalized(), tampered); err == nil {
		t.Fatal("tampered checkpoint verified")
	}
}

func TestPublicSnapshotAndCheckpointEndpoints(t *testing.T) {
	f := newExecutionFixture(t, 1)
	server := &HTTPServer{
		NetworkID: f.genesis.NetworkID,
		Directory: NewDirectory(f.genesis.NetworkID, 16, 4),
		Mempool:   NewMempool(f.genesis.NetworkID, 16),
		Ledger:    f.states[0],
		Height:    f.engines[0].Height,
		Checkpoint: func() (Checkpoint, error) {
			return BuildCheckpoint(f.engines[0], f.states[0])
		},
	}
	for _, path := range []string{"/v1/public/snapshot", "/v1/public/checkpoint"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s status=%d body=%s", path, rec.Code, rec.Body.String())
		}
	}
}
