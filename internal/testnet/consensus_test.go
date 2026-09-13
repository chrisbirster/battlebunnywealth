package testnet

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type fixture struct {
	genesis    Genesis
	validators []Validator
	keys       []any
	now        time.Time
}

func makeFixture(t *testing.T) fixture {
	t.Helper()
	now := time.Unix(1_800_000_000, 0).UTC()
	validators := make([]Validator, 4)
	keys := make([]any, 4)
	for i := 0; i < 4; i++ {
		pub, priv, err := GenerateValidatorKey(AlgorithmEd25519)
		if err != nil {
			t.Fatal(err)
		}
		validators[i] = Validator{ID: fmt.Sprintf("validator-%02d", i+1), AccountID: fmt.Sprintf("a%d", i), DeviceID: fmt.Sprintf("d%d", i), Algorithm: AlgorithmEd25519, PublicKey: pub, Authority: 1000}
		keys[i] = priv
	}
	g, err := NewGenesis("testnet", now, DefaultProtocolConfig(), validators)
	if err != nil {
		t.Fatal(err)
	}
	return fixture{g, g.Validators, keys, now}
}
func (f fixture) proposal(t *testing.T, e *Engine, round uint32) (Block, Proposal) {
	t.Helper()
	committee := SelectCommittee(f.validators, f.genesis.Config.CommitteeTarget, e.Status().FinalizedHash, e.Height()+1, round)
	p, _ := ExpectedProposer(committee, e.Height()+1, round)
	idx := 0
	for i, v := range f.validators {
		if v.ID == p.ID {
			idx = i
		}
	}
	b, err := e.Draft(round, p.ID, nil, f.now.Add(time.Duration(e.Height()+1)*time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	proposal, err := BuildProposal(b, p, f.keys[idx])
	if err != nil {
		t.Fatal(err)
	}
	return b, proposal
}
func TestFourValidatorBootstrapFinalizesWithThree(t *testing.T) {
	f := makeFixture(t)
	engines := make([]*Engine, 5)
	for i := range engines {
		engines[i], _ = NewEngine(f.genesis)
	}
	b, p := f.proposal(t, engines[0], 0)
	for _, e := range engines {
		if err := e.HandleProposal(p); err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 3; i++ {
		v, err := BuildVote(b, f.validators[i], f.keys[i])
		if err != nil {
			t.Fatal(err)
		}
		for _, e := range engines {
			_, err = e.HandleVote(v)
			if err != nil {
				t.Fatal(err)
			}
		}
	}
	for i, e := range engines {
		if e.Height() != 1 {
			t.Fatalf("node %d height=%d", i, e.Height())
		}
	}
}
func TestTwoOfflineStopsFinality(t *testing.T) {
	f := makeFixture(t)
	e, _ := NewEngine(f.genesis)
	b, p := f.proposal(t, e, 0)
	if err := e.HandleProposal(p); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		v, _ := BuildVote(b, f.validators[i], f.keys[i])
		done, err := e.HandleVote(v)
		if err != nil {
			t.Fatal(err)
		}
		if done {
			t.Fatal("finalized without 3-of-4 quorum")
		}
	}
	if e.Height() != 0 {
		t.Fatal("height advanced")
	}
}
func TestInvalidSignatureAndEquivocationRejected(t *testing.T) {
	f := makeFixture(t)
	e, _ := NewEngine(f.genesis)
	b, p := f.proposal(t, e, 0)
	if err := e.HandleProposal(p); err != nil {
		t.Fatal(err)
	}
	vote, _ := BuildVote(b, f.validators[0], f.keys[0])
	vote.Signature = "bogus"
	if _, err := e.HandleVote(vote); !errors.Is(err, ErrInvalidSignature) {
		t.Fatalf("got %v", err)
	}
	vote, _ = BuildVote(b, f.validators[0], f.keys[0])
	if _, err := e.HandleVote(vote); err != nil {
		t.Fatal(err)
	}
	other := vote
	other.BlockHash = "different"
	sig, _ := Sign(f.validators[0].Algorithm, f.keys[0], voteSigningMessage(other))
	other.Signature = sig
	if _, err := e.HandleVote(other); !errors.Is(err, ErrEquivocation) {
		t.Fatalf("got %v", err)
	}
}
func TestHundredFakeNodesHaveZeroConsensusPower(t *testing.T) {
	f := makeFixture(t)
	e, _ := NewEngine(f.genesis)
	b, p := f.proposal(t, e, 0)
	if err := e.HandleProposal(p); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 100; i++ {
		pub, priv, _ := GenerateValidatorKey(AlgorithmEd25519)
		fake := Validator{ID: validatorID(pub), Algorithm: AlgorithmEd25519, PublicKey: pub}
		v, _ := BuildVote(b, fake, priv)
		if _, err := e.HandleVote(v); !errors.Is(err, ErrUnknownValidator) {
			t.Fatalf("fake node %d got %v", i, err)
		}
	}
	if e.Height() != 0 {
		t.Fatal("fake nodes finalized block")
	}
}
func TestActivationRateLimitsPhoneFarm(t *testing.T) {
	f := makeFixture(t)
	r := NewActivationRegistry(DefaultActivationPolicy(), f.genesis.Validators)
	for i := 0; i < 100; i++ {
		pub, _, _ := GenerateValidatorKey(AlgorithmEd25519)
		r.AddCandidate(ValidatorCandidate{Validator: Validator{ID: validatorID(pub), Algorithm: AlgorithmEd25519, PublicKey: pub, Authority: 1000}, Provider: "google-play-integrity", Attested: true, JoinedAt: f.now})
	}
	if got := r.ActivateEligible(f.now.Add(29 * 24 * time.Hour)); len(got) != 0 {
		t.Fatalf("activated before maturity: %d", len(got))
	}
	if got := r.ActivateEligible(f.now.Add(30 * 24 * time.Hour)); len(got) != 1 {
		t.Fatalf("first window activated %d", len(got))
	}
	if got := r.ActivateEligible(f.now.Add(31 * 24 * time.Hour)); len(got) != 0 {
		t.Fatalf("same window activated %d", len(got))
	}
	if got := r.ActivateEligible(f.now.Add(37 * 24 * time.Hour)); len(got) != 1 {
		t.Fatalf("second window activated %d", len(got))
	}
}
func TestPeerAllowlistRejectsUnknownAndReplay(t *testing.T) {
	f := makeFixture(t)
	key, _ := GenerateNodeKey()
	guard, err := NewPeerGuard(f.genesis.NetworkID, 5*time.Minute, []Peer{{NodeID: key.ID(), PublicKey: key.PublicText()}})
	if err != nil {
		t.Fatal(err)
	}
	payload := map[string]string{"hello": "world"}
	env, _ := SignPeerEnvelope(f.genesis.NetworkID, "proposal", "n1", key, payload, f.now)
	if err := guard.Verify(env, f.now); err != nil {
		t.Fatal(err)
	}
	if err := guard.Verify(env, f.now); !errors.Is(err, ErrReplay) {
		t.Fatalf("got %v", err)
	}
	evilPub, evilPriv, _ := ed25519.GenerateKey(nil)
	evil := NodeKey{Public: evilPub, Private: evilPriv}
	env2, _ := SignPeerEnvelope(f.genesis.NetworkID, "proposal", "n2", evil, payload, f.now)
	if err := guard.Verify(env2, f.now); !errors.Is(err, ErrUnknownPeer) {
		t.Fatalf("got %v", err)
	}
}
func TestHTTPStatus(t *testing.T) {
	f := makeFixture(t)
	e, _ := NewEngine(f.genesis)
	node := NewHTTPNode(e, nil, nil)
	req := httptest.NewRequest("GET", "/v1/node/status", nil)
	rec := httptest.NewRecorder()
	node.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}
func TestP256ValidatorSignature(t *testing.T) {
	pub, priv, err := GenerateValidatorKey(AlgorithmP256SPKI)
	if err != nil {
		t.Fatal(err)
	}
	sig, err := Sign(AlgorithmP256SPKI, priv, "hello")
	if err != nil {
		t.Fatal(err)
	}
	if !VerifySignature(AlgorithmP256SPKI, pub, "hello", sig) {
		t.Fatal("signature failed")
	}
	raw, _ := base64.RawURLEncoding.DecodeString(pub)
	if len(raw) == 0 {
		t.Fatal("missing spki")
	}
}
func TestSmokeCluster(t *testing.T) {
	if err := RunSmokeCluster(10); err != nil {
		t.Fatal(err)
	}
}
func TestDuplicateVoteRejected(t *testing.T) {
	f := makeFixture(t)
	e, _ := NewEngine(f.genesis)
	b, p := f.proposal(t, e, 0)
	if err := e.HandleProposal(p); err != nil {
		t.Fatal(err)
	}
	v, _ := BuildVote(b, f.validators[0], f.keys[0])
	if _, err := e.HandleVote(v); err != nil {
		t.Fatal(err)
	}
	if _, err := e.HandleVote(v); !errors.Is(err, ErrDuplicateVote) {
		t.Fatalf("got %v", err)
	}
}
func TestProposalRejectsWrongNetworkAndPreviousHash(t *testing.T) {
	f := makeFixture(t)
	e, _ := NewEngine(f.genesis)
	_, p := f.proposal(t, e, 0)
	wrongNetwork := p
	wrongNetwork.Block.NetworkID = "evil-net"
	wrongNetwork.Block.Hash = hashBlock(wrongNetwork.Block)
	if err := e.HandleProposal(wrongNetwork); !errors.Is(err, ErrWrongNetwork) {
		t.Fatalf("network err=%v", err)
	}
	wrongPrev := p
	wrongPrev.Block.PreviousHash = "deadbeef"
	wrongPrev.Block.Hash = hashBlock(wrongPrev.Block)
	if err := e.HandleProposal(wrongPrev); !errors.Is(err, ErrInvalidPreviousHash) {
		t.Fatalf("previous err=%v", err)
	}
}
func TestStoreRestartAndPeerCatchup(t *testing.T) {
	f := makeFixture(t)
	e, _ := NewEngine(f.genesis)
	b, p := f.proposal(t, e, 0)
	if err := e.HandleProposal(p); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		v, _ := BuildVote(b, f.validators[i], f.keys[i])
		if _, err := e.HandleVote(v); err != nil {
			t.Fatal(err)
		}
	}
	dir := t.TempDir()
	store, err := OpenStore(dir, f.genesis)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Append(e.Finalized()[0]); err != nil {
		t.Fatal(err)
	}
	restarted, _ := NewEngine(f.genesis)
	if err := store.Restore(restarted); err != nil {
		t.Fatal(err)
	}
	if restarted.Status().FinalizedHash != e.Status().FinalizedHash {
		t.Fatal("restart did not recover finalized head")
	}
	server := httptest.NewServer(NewHTTPNode(e, nil, nil))
	defer server.Close()
	stale, _ := NewEngine(f.genesis)
	if err := SyncFrom(server.URL, server.Client(), stale); err != nil {
		t.Fatal(err)
	}
	if stale.Height() != 1 || stale.Status().FinalizedHash != e.Status().FinalizedHash {
		t.Fatal("stale peer did not catch up")
	}
}
func TestDirectIngressRelaysToPermissionedPeer(t *testing.T) {
	f := makeFixture(t)
	engineA, _ := NewEngine(f.genesis)
	engineB, _ := NewEngine(f.genesis)
	keyA, _ := GenerateNodeKey()
	keyB, _ := GenerateNodeKey()
	guardB, err := NewPeerGuard(f.genesis.NetworkID, 5*time.Minute, []Peer{{NodeID: keyA.ID(), PublicKey: keyA.PublicText(), URL: "http://a"}})
	if err != nil {
		t.Fatal(err)
	}
	nodeB := NewHTTPNode(engineB, guardB, nil)
	serverB := httptest.NewServer(nodeB)
	defer serverB.Close()
	guardA, err := NewPeerGuard(f.genesis.NetworkID, 5*time.Minute, []Peer{{NodeID: keyB.ID(), PublicKey: keyB.PublicText(), URL: serverB.URL}})
	if err != nil {
		t.Fatal(err)
	}
	nodeA := NewHTTPNode(engineA, guardA, nil)
	nodeA.Relayer = &Relayer{NetworkID: f.genesis.NetworkID, Key: keyA, Peers: []Peer{{NodeID: keyB.ID(), PublicKey: keyB.PublicText(), URL: serverB.URL}}, Client: serverB.Client()}
	serverA := httptest.NewServer(nodeA)
	defer serverA.Close()
	b, p := f.proposal(t, engineA, 0)
	raw, _ := json.Marshal(p)
	req, _ := http.NewRequest(http.MethodPost, serverA.URL+"/v1/committee/proposals", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	resp, err := serverA.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("proposal status=%d", resp.StatusCode)
	}
	if engineB.Status().PendingProposal != b.Hash {
		t.Fatalf("peer did not receive proposal: %+v", engineB.Status())
	}
}
