package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/chrisbirster/battlebunnywealth/internal/testnet"
)

func main() {
	blocks := flag.Int("blocks", 20, "number of finalized blocks in the deterministic local cluster")
	flag.Parse()
	if err := testnet.RunSmokeCluster(*blocks); err != nil {
		fmt.Fprintln(os.Stderr, "pop-cluster:", err)
		os.Exit(1)
	}
	fmt.Printf("pop-cluster: %d blocks finalized across 5 independent engines\n", *blocks)
}
