package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/chrisbirster/battlebunnywealth/internal/publictestnet"
)

func main() {
	blocks := flag.Int("blocks", 1100, "number of finalized blocks to execute")
	gate := flag.Bool("gate", false, "exit non-zero unless transition/recovery expectations are satisfied")
	flag.Parse()

	report, err := publictestnet.RunSoak(*blocks)
	if err != nil {
		fatal(err)
	}
	_ = json.NewEncoder(os.Stdout).Encode(report)
	if *gate {
		if report.FinalHeight != uint64(report.Blocks) || report.Restarts == 0 || report.TemporaryPartitions == 0 || report.AlternateQuorumRounds == 0 {
			fatal(fmt.Errorf("public-testnet soak gate failed"))
		}
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "pop-soak:", err)
	os.Exit(1)
}
