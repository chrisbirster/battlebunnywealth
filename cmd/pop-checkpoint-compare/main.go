package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/chrisbirster/battlebunnywealth/internal/publictestnet"
)

type endpointsFlag []string

func (e *endpointsFlag) String() string { return strings.Join(*e, ",") }
func (e *endpointsFlag) Set(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("empty endpoint")
	}
	*e = append(*e, value)
	return nil
}

func main() {
	var endpoints endpointsFlag
	flag.Var(&endpoints, "url", "public-testnet node base URL; repeat at least twice")
	timeout := flag.Duration("timeout", 10*time.Second, "per-request timeout")
	flag.Parse()
	if len(endpoints) < 2 {
		fatal(fmt.Errorf("at least two -url values are required"))
	}
	client := &http.Client{Timeout: *timeout}
	observations := make([]publictestnet.CheckpointObservation, 0, len(endpoints))
	for _, endpoint := range endpoints {
		checkpoint, err := fetchCheckpoint(client, endpoint)
		if err != nil {
			fatal(fmt.Errorf("%s: %w", endpoint, err))
		}
		observations = append(observations, publictestnet.CheckpointObservation{Source: endpoint, Checkpoint: checkpoint})
	}
	comparison, err := publictestnet.CompareCheckpointSet(observations)
	_ = json.NewEncoder(os.Stdout).Encode(comparison)
	if err != nil {
		fatal(err)
	}
}

func fetchCheckpoint(client *http.Client, base string) (publictestnet.Checkpoint, error) {
	url := strings.TrimRight(base, "/") + "/v1/public/checkpoint"
	resp, err := client.Get(url)
	if err != nil {
		return publictestnet.Checkpoint{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return publictestnet.Checkpoint{}, fmt.Errorf("status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var checkpoint publictestnet.Checkpoint
	dec := json.NewDecoder(io.LimitReader(resp.Body, 1<<20))
	if err := dec.Decode(&checkpoint); err != nil {
		return publictestnet.Checkpoint{}, err
	}
	return checkpoint, nil
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "pop-checkpoint-compare:", err)
	os.Exit(1)
}
