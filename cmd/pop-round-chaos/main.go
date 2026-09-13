package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/chrisbirster/battlebunnywealth/internal/publictestnet"
)

func main() {
	blocks := flag.Int("blocks", 240, "number of finalized blocks")
	seed := flag.Int64("seed", 42, "deterministic chaos seed")
	gate := flag.Bool("gate", false, "exit non-zero when required chaos coverage is missing")
	flag.Parse()

	report, err := publictestnet.RunRoundChaos(*blocks, *seed)
	if err != nil {
		fatal(err)
	}
	if err := json.NewEncoder(os.Stdout).Encode(report); err != nil {
		fatal(err)
	}
	if *gate {
		if report.FinalHeight != uint64(*blocks) || report.RoundChanges == 0 || report.LockedRoundChanges == 0 || report.Partitions == 0 || report.Restarts == 0 || report.DuplicateMessages == 0 || report.ReorderedDeliveries == 0 {
			fatal(fmt.Errorf("round chaos gate incomplete: %+v", report))
		}
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "pop-round-chaos:", err)
	os.Exit(1)
}
