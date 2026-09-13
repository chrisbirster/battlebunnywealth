package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
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

type nodeEvidence struct {
	URL                 string                        `json:"url"`
	Available           bool                          `json:"available"`
	StatusLatencyMillis int64                         `json:"statusLatencyMillis"`
	CheckpointLatencyMillis int64                     `json:"checkpointLatencyMillis"`
	Status              map[string]any                `json:"status,omitempty"`
	Checkpoint          *publictestnet.Checkpoint     `json:"checkpoint,omitempty"`
	Error               string                        `json:"error,omitempty"`
}

type evidenceReport struct {
	Version          int                             `json:"version"`
	CollectedAt      time.Time                       `json:"collectedAt"`
	Nodes            []nodeEvidence                  `json:"nodes"`
	Available        int                             `json:"available"`
	Unavailable      int                             `json:"unavailable"`
	CommonHeight     uint64                          `json:"commonHeight,omitempty"`
	CheckpointMatch  *bool                           `json:"checkpointMatch,omitempty"`
	Comparison       *publictestnet.CheckpointComparison `json:"comparison,omitempty"`
	ComparisonError  string                          `json:"comparisonError,omitempty"`
}

func main() {
	var endpoints endpointsFlag
	flag.Var(&endpoints, "url", "public-testnet node base URL; repeat for every operator")
	timeout := flag.Duration("timeout", 10*time.Second, "per-request timeout")
	minAvailable := flag.Int("min-available", 2, "minimum nodes that must answer when -gate is set")
	gate := flag.Bool("gate", false, "exit non-zero on availability or same-height checkpoint disagreement")
	flag.Parse()
	if len(endpoints) == 0 {
		fatal(fmt.Errorf("at least one -url is required"))
	}

	client := &http.Client{Timeout: *timeout}
	report := evidenceReport{Version: 1, CollectedAt: time.Now().UTC()}
	for _, endpoint := range endpoints {
		report.Nodes = append(report.Nodes, collectNode(client, endpoint))
	}
	sort.Slice(report.Nodes, func(i, j int) bool { return report.Nodes[i].URL < report.Nodes[j].URL })
	for _, node := range report.Nodes {
		if node.Available {
			report.Available++
		} else {
			report.Unavailable++
		}
	}

	observations := sameHeightObservations(report.Nodes)
	if len(observations) >= 2 {
		report.CommonHeight = observations[0].Checkpoint.Height
		comparison, err := publictestnet.CompareCheckpointSet(observations)
		match := err == nil
		report.CheckpointMatch = &match
		report.Comparison = &comparison
		if err != nil {
			report.ComparisonError = err.Error()
		}
	}

	_ = json.NewEncoder(os.Stdout).Encode(report)
	if *gate {
		if report.Available < *minAvailable {
			fatal(fmt.Errorf("distributed evidence gate: available=%d want>=%d", report.Available, *minAvailable))
		}
		if report.CheckpointMatch != nil && !*report.CheckpointMatch {
			fatal(fmt.Errorf("distributed evidence gate: %s", report.ComparisonError))
		}
	}
}

func collectNode(client *http.Client, base string) nodeEvidence {
	base = strings.TrimRight(base, "/")
	result := nodeEvidence{URL: base}
	started := time.Now()
	status, err := getJSONMap(client, base+"/v1/public/status")
	result.StatusLatencyMillis = time.Since(started).Milliseconds()
	if err != nil {
		result.Error = "status: " + err.Error()
		return result
	}
	result.Status = status
	started = time.Now()
	checkpoint, err := getCheckpoint(client, base+"/v1/public/checkpoint")
	result.CheckpointLatencyMillis = time.Since(started).Milliseconds()
	if err != nil {
		result.Error = "checkpoint: " + err.Error()
		return result
	}
	result.Checkpoint = &checkpoint
	result.Available = true
	return result
}

func getJSONMap(client *http.Client, url string) (map[string]any, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var value map[string]any
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&value); err != nil {
		return nil, err
	}
	return value, nil
}

func getCheckpoint(client *http.Client, url string) (publictestnet.Checkpoint, error) {
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
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&checkpoint); err != nil {
		return publictestnet.Checkpoint{}, err
	}
	return checkpoint, nil
}

func sameHeightObservations(nodes []nodeEvidence) []publictestnet.CheckpointObservation {
	counts := map[uint64]int{}
	for _, node := range nodes {
		if node.Available && node.Checkpoint != nil {
			counts[node.Checkpoint.Height]++
		}
	}
	var selected uint64
	best := 0
	for height, count := range counts {
		if count > best || (count == best && height > selected) {
			selected, best = height, count
		}
	}
	if best < 2 {
		return nil
	}
	out := []publictestnet.CheckpointObservation{}
	for _, node := range nodes {
		if node.Available && node.Checkpoint != nil && node.Checkpoint.Height == selected {
			out = append(out, publictestnet.CheckpointObservation{Source: node.URL, Checkpoint: *node.Checkpoint})
		}
	}
	return out
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "pop-network-evidence:", err)
	os.Exit(1)
}
