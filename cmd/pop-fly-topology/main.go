package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/chrisbirster/battlebunnywealth/internal/publictestnet"
)

type topologySource struct {
	OperatorID   string `json:"operatorId"`
	AppName      string `json:"appName"`
	Region       string `json:"region"`
	MachinesFile string `json:"machinesFile"`
}

func main() {
	inventoryPath := flag.String("inventory", "", "JSON file describing four Fly apps and their captured machines-list files")
	repositorySHA := flag.String("repo-sha", "", "exact merged dev commit SHA deployed to the Fly apps")
	outPath := flag.String("out", "fly-topology.json", "output topology evidence JSON path")
	flag.Parse()
	if strings.TrimSpace(*inventoryPath) == "" || strings.TrimSpace(*repositorySHA) == "" {
		fatal(errors.New("-inventory and -repo-sha are required"))
	}

	raw, err := os.ReadFile(*inventoryPath)
	if err != nil {
		fatal(err)
	}
	var sources []topologySource
	if err := json.Unmarshal(raw, &sources); err != nil {
		fatal(fmt.Errorf("decode inventory: %w", err))
	}
	inputs := make([]publictestnet.NamedFlyMachineList, 0, len(sources))
	for _, source := range sources {
		if strings.TrimSpace(source.MachinesFile) == "" {
			fatal(errors.New("every inventory entry requires machinesFile"))
		}
		machines, err := os.ReadFile(source.MachinesFile)
		if err != nil {
			fatal(fmt.Errorf("read %s: %w", source.MachinesFile, err))
		}
		inputs = append(inputs, publictestnet.NamedFlyMachineList{
			OperatorID:     source.OperatorID,
			AppName:        source.AppName,
			ExpectedRegion: source.Region,
			Raw:            machines,
		})
	}

	evidence, err := publictestnet.BuildFlyTopologyEvidence(strings.TrimSpace(*repositorySHA), time.Now().UTC(), inputs)
	if err != nil {
		fatal(err)
	}
	payload, err := json.MarshalIndent(evidence, "", "  ")
	if err != nil {
		fatal(err)
	}
	if err := os.WriteFile(*outPath, append(payload, '\n'), 0o600); err != nil {
		fatal(err)
	}
	fmt.Printf("fly topology: %s nodes=%d images=%d hash=%s\n", *outPath, len(evidence.Nodes), len(evidence.ImageDigests), evidence.Hash)
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "pop-fly-topology:", err)
	os.Exit(1)
}
