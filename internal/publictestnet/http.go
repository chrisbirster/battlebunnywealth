package publictestnet

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/chrisbirster/battlebunnywealth/internal/carrot"
)

type LedgerView interface {
	Balance(string) int64
	Nonce(string) uint64
	SupplyReport() carrot.SupplyReport
}

type transitionView interface{ TransitionStatus() TransitionStatus }
type snapshotView interface{ Snapshot() ConsensusSnapshot }

type HTTPServer struct {
	Base http.Handler
	NetworkID string
	Directory *Directory
	Mempool *Mempool
	Ledger LedgerView
	Admission *AdmissionService
	Height func() uint64
	Checkpoint func() (Checkpoint, error)
}

func (s *HTTPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/v1/public/status":
		s.handleStatus(w)
	case r.Method == http.MethodGet && r.URL.Path == "/v1/public/funding-policy":
		s.writeJSON(w, http.StatusOK, DefaultTestFundingPolicy())
	case r.Method == http.MethodGet && r.URL.Path == "/v1/public/snapshot":
		view, ok := s.Ledger.(snapshotView)
		if !ok { http.Error(w, "consensus snapshot unavailable", http.StatusServiceUnavailable); return }
		s.writeJSON(w, http.StatusOK, view.Snapshot())
	case r.Method == http.MethodGet && r.URL.Path == "/v1/public/checkpoint":
		if s.Checkpoint == nil { http.Error(w, "checkpoint unavailable", http.StatusServiceUnavailable); return }
		checkpoint, err := s.Checkpoint()
		if err != nil { http.Error(w, "checkpoint unavailable", http.StatusServiceUnavailable); return }
		s.writeJSON(w, http.StatusOK, checkpoint)
	case r.Method == http.MethodGet && r.URL.Path == "/v1/public/peers":
		s.writeJSON(w, http.StatusOK, s.Directory.List(time.Now().UTC()))
	case r.Method == http.MethodPost && r.URL.Path == "/v1/public/peers":
		var a NodeAnnouncement
		if !s.decode(w, r, &a) { return }
		if err := s.Directory.Upsert(a, time.Now().UTC()); err != nil { http.Error(w, err.Error(), http.StatusBadRequest); return }
		s.writeJSON(w, http.StatusAccepted, map[string]any{"accepted": true})
	case r.Method == http.MethodGet && r.URL.Path == "/v1/public/mempool":
		s.writeJSON(w, http.StatusOK, s.Mempool.Transactions())
	case r.Method == http.MethodPost && r.URL.Path == "/v1/public/transactions":
		var tx SignedTransaction
		if !s.decode(w, r, &tx) { return }
		if err := s.Mempool.Admit(tx, s.currentHeight()+1, s.Ledger.Nonce(tx.From)); err != nil { http.Error(w, err.Error(), http.StatusBadRequest); return }
		s.writeJSON(w, http.StatusAccepted, map[string]any{"accepted": true, "id": tx.ID})
	case r.Method == http.MethodGet && r.URL.Path == "/v1/public/ledger":
		account := r.URL.Query().Get("account")
		s.writeJSON(w, http.StatusOK, map[string]any{"account": account, "balanceAtoms": s.Ledger.Balance(account), "nonce": s.Ledger.Nonce(account), "supply": s.Ledger.SupplyReport()})
	case r.Method == http.MethodPost && r.URL.Path == "/v1/public/validators/apply":
		if s.Admission == nil { http.Error(w, "validator admission unavailable", http.StatusServiceUnavailable); return }
		var a CandidateApplication
		if !s.decode(w, r, &a) { return }
		if err := s.Admission.Submit(r.Context(), a, time.Now().UTC()); err != nil { http.Error(w, err.Error(), http.StatusBadRequest); return }
		s.writeJSON(w, http.StatusAccepted, map[string]any{"accepted": true, "pending": s.Admission.Pending()})
	default:
		if s.Base != nil { s.Base.ServeHTTP(w, r); return }
		http.NotFound(w, r)
	}
}

func (s *HTTPServer) currentHeight() uint64 { if s.Height == nil { return 0 }; return s.Height() }

func (s *HTTPServer) handleStatus(w http.ResponseWriter) {
	peers := 0
	if s.Directory != nil { peers = len(s.Directory.List(time.Now().UTC())) }
	pending, active := 0, 0
	if s.Admission != nil { pending = s.Admission.Pending(); active = len(s.Admission.Active()) }
	mempoolCount := 0
	if s.Mempool != nil { mempoolCount = s.Mempool.Count() }
	funding := DefaultTestFundingPolicy()
	status := map[string]any{"networkId": s.NetworkID, "height": s.currentHeight(), "publicPeers": peers, "mempool": mempoolCount, "validatorCandidates": pending, "activeValidators": active, "carrotPolicyHash": carrot.DefaultPolicy().Hash(), "testFundingPolicyHash": funding.Hash, "asset": "TEST-CARROT", "economicValue": false, "consensusExecution": true}
	if view, ok := s.Ledger.(transitionView); ok { transitions := view.TransitionStatus(); status["transitions"] = transitions; status["activeValidators"] = transitions.ActiveValidatorCount; status["protocolVersion"] = transitions.ActiveProtocolVersion }
	s.writeJSON(w, http.StatusOK, status)
}

func (s *HTTPServer) decode(w http.ResponseWriter, r *http.Request, out any) bool {
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") { http.Error(w, "content-type must be application/json", http.StatusUnsupportedMediaType); return false }
	r.Body = http.MaxBytesReader(w, r.Body, 128<<10)
	dec := json.NewDecoder(r.Body); dec.DisallowUnknownFields()
	if err := dec.Decode(out); err != nil { http.Error(w, "invalid json", http.StatusBadRequest); return false }
	if err := dec.Decode(&struct{}{}); err != io.EOF { http.Error(w, "multiple json values", http.StatusBadRequest); return false }
	return true
}

func (s *HTTPServer) writeJSON(w http.ResponseWriter, status int, v any) { w.Header().Set("Content-Type", "application/json"); w.WriteHeader(status); _ = json.NewEncoder(w).Encode(v) }
