package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/chrisbirster/battlebunnywealth/internal/carrot"
)

type output struct {
	Policy       carrot.Policy       `json:"policy"`
	PolicyHash   string              `json:"policyHash"`
	Height       uint64              `json:"height"`
	RewardAtoms  int64               `json:"rewardAtoms"`
	RewardCARROT string              `json:"rewardCarrot"`
	Supply       carrot.SupplyReport `json:"supply"`
}

func main() {
	height := flag.Uint64("height", 1, "finalized height to inspect")
	check := flag.Bool("check", false, "validate frozen CARROT policy invariants")
	flag.Parse()

	policy := carrot.DefaultPolicy()
	if err := policy.Validate(); err != nil {
		fail(err)
	}
	if carrot.TotalScheduledParticipation() != carrot.ParticipationReserveAtoms {
		fail(fmt.Errorf("participation schedule does not exhaust reserve"))
	}
	ledger, err := carrot.NewLedger(policy)
	if err != nil {
		fail(err)
	}
	if *check {
		if err := ledger.ValidateConservation(); err != nil {
			fail(err)
		}
		fmt.Printf("carrot-spec: ok policy=%s max=%s founder=20%% participation=60%%\n", policy.Hash(), carrot.FormatAtoms(carrot.MaxSupplyAtoms))
		return
	}
	for h := uint64(1); h <= *height; h++ {
		if _, err := ledger.ReleaseParticipation(h, []string{"testnet-finality-set"}); err != nil {
			fail(err)
		}
	}
	reward := carrot.RewardAtHeight(*height)
	out := output{Policy: policy, PolicyHash: policy.Hash(), Height: *height, RewardAtoms: reward, RewardCARROT: carrot.FormatAtoms(reward), Supply: ledger.SupplyReport()}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		fail(err)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "carrot-spec:", err)
	os.Exit(1)
}
