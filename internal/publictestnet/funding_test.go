package publictestnet

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDefaultTestFundingPolicyIsNonEconomicAndNonMinting(t *testing.T) {
	p := DefaultTestFundingPolicy()
	if err := p.Validate(); err != nil {
		t.Fatal(err)
	}
	if p.EconomicValue || p.AutomaticFaucetEnabled || p.TreasurySpending {
		t.Fatal("test funding policy enabled an economic or treasury path")
	}
	if p.Hash == "" {
		t.Fatal("missing funding policy hash")
	}
}

func TestFundingPolicyEndpoint(t *testing.T) {
	s := &HTTPServer{NetworkID: "public"}
	req := httptest.NewRequest(http.MethodGet, "/v1/public/funding-policy", nil)
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var got TestFundingPolicy
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if err := got.Validate(); err != nil {
		t.Fatal(err)
	}
}
