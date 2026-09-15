package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/chrisbirster/battlebunnywealth/internal/testnet"
)

type summary struct {
	NetworkID   string            `json:"networkId"`
	GenesisHash string            `json:"genesisHash"`
	CreatedAt   time.Time         `json:"createdAt"`
	Files       map[string]string `json:"files"`
}

func main() {
	validatorsPath := flag.String("validators", "", "JSON file containing exactly four genesis validators")
	nodesPath := flag.String("nodes", "", "JSON file containing exactly four public full-node identities/endpoints")
	outDir := flag.String("out", ".fly-private/bootstrap", "output directory")
	networkID := flag.String("network-id", "bbw-pop-fly-testnet-v1", "testnet network id")
	createdAtRaw := flag.String("created-at", "", "genesis RFC3339 timestamp; defaults to current UTC time")
	syncSeconds := flag.Int("sync-interval", 5, "peer catch-up interval in seconds")
	flag.Parse()
	if *validatorsPath == "" || *nodesPath == "" {
		fatal(errors.New("-validators and -nodes are required"))
	}

	var validators []testnet.Validator
	readJSON(*validatorsPath, &validators)
	if len(validators) != 4 { fatal(fmt.Errorf("exactly four genesis validators required, got %d", len(validators))) }
	var nodes []testnet.BootstrapNode
	readJSON(*nodesPath, &nodes)
	if len(nodes) != 4 { fatal(fmt.Errorf("exactly four full nodes required, got %d", len(nodes))) }

	createdAt := time.Now().UTC().Truncate(time.Second)
	if *createdAtRaw != "" {
		parsed, err := time.Parse(time.RFC3339, *createdAtRaw)
		if err != nil { fatal(fmt.Errorf("parse -created-at: %w", err)) }
		createdAt = parsed.UTC()
	}
	genesis, err := testnet.NewGenesis(*networkID, createdAt, testnet.DefaultProtocolConfig(), validators)
	if err != nil { fatal(err) }
	configs, err := testnet.BuildNodeConfigs(genesis, nodes, *syncSeconds)
	if err != nil { fatal(err) }
	if err := os.MkdirAll(*outDir, 0o700); err != nil { fatal(err) }

	files := map[string]string{}
	genesisPath := filepath.Join(*outDir, "genesis.json")
	writeJSON(genesisPath, genesis)
	files["genesis"] = genesisPath

	names := make([]string, 0, len(configs))
	for name := range configs { names = append(names, name) }
	sort.Strings(names)
	for _, name := range names {
		path := filepath.Join(*outDir, name+".node.json")
		writeJSON(path, configs[name])
		files[name] = path
	}
	result := summary{NetworkID: genesis.NetworkID, GenesisHash: genesis.Hash, CreatedAt: genesis.CreatedAt, Files: files}
	writeJSON(filepath.Join(*outDir, "bootstrap-summary.json"), result)
	_ = json.NewEncoder(os.Stdout).Encode(result)
}

func readJSON(path string, target any) {
	raw, err := os.ReadFile(path)
	if err != nil { fatal(err) }
	if err := json.Unmarshal(raw, target); err != nil { fatal(fmt.Errorf("decode %s: %w", path, err)) }
}

func writeJSON(path string, value any) {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil { fatal(err) }
	raw = append(raw, '\n')
	if err := os.WriteFile(path, raw, 0o600); err != nil { fatal(err) }
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "pop-bootstrap-config:", err)
	os.Exit(1)
}
