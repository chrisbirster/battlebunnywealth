package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/chrisbirster/battlebunnywealth/internal/publictestnet"
)

func main() {
	operatorsPath := flag.String("operators", "", "JSON file containing public-testnet operator descriptors")
	repositorySHA := flag.String("repo-sha", "", "exact repository commit SHA running on the operators")
	networkMapHash := flag.String("network-map-hash", "", "optional canonical peer network-map hash")
	samples := flag.Int("samples", 1, "number of collection cycles")
	interval := flag.Duration("interval", 30*time.Second, "delay between collection cycles")
	timeout := flag.Duration("timeout", 10*time.Second, "per-request timeout")
	outPath := flag.String("out", "", "output evidence JSON path; stdout when empty")
	gate := flag.Bool("gate", false, "exit non-zero when review-readiness policy is not met")
	minProviders := flag.Int("min-providers", 2, "minimum distinct declared providers when gating")
	minRegions := flag.Int("min-regions", 1, "minimum distinct declared regions when gating")
	minSamples := flag.Int("min-samples-per-operator", 1, "minimum available samples per operator when gating")
	minWindow := flag.Duration("min-window", 0, "minimum evidence duration when gating")
	flag.Parse()
	if *operatorsPath == "" || strings.TrimSpace(*repositorySHA) == "" {
		fatal(errors.New("-operators and -repo-sha are required"))
	}
	if *samples <= 0 {
		fatal(errors.New("-samples must be positive"))
	}
	operators, err := loadOperators(*operatorsPath)
	if err != nil {
		fatal(err)
	}
	if len(operators) < 2 {
		fatal(errors.New("at least two operators are required"))
	}

	client := &http.Client{Timeout: *timeout}
	window := publictestnet.LiveEvidenceWindow{Version: publictestnet.LiveEvidenceVersion, RepositorySHA: strings.TrimSpace(*repositorySHA), NetworkMapHash: strings.TrimSpace(*networkMapHash), StartedAt: time.Now().UTC(), Operators: operators}
	for sequence := 1; sequence <= *samples; sequence++ {
		for _, operator := range operators {
			window.Samples = append(window.Samples, collect(client, sequence, operator))
		}
		if sequence < *samples {
			time.Sleep(*interval)
		}
	}
	window.EndedAt = time.Now().UTC()
	for _, sample := range window.Samples {
		if sample.Checkpoint == nil {
			continue
		}
		window.NetworkID = sample.Checkpoint.NetworkID
		window.GenesisHash = sample.Checkpoint.GenesisHash
		window.CarrotPolicyHash = sample.Checkpoint.CarrotPolicyHash
		break
	}
	if window.NetworkID == "" {
		fatal(errors.New("no operator returned a checkpoint; evidence commitments cannot be established"))
	}
	if err := window.Finalize(); err != nil {
		fatal(err)
	}
	readiness := window.EvaluateReadiness(publictestnet.EvidenceReadinessPolicy{MinOperators: len(operators), MinProviders: *minProviders, MinRegions: *minRegions, MinSamplesPerOperator: *minSamples, MinWindow: *minWindow})
	payload := struct {
		Evidence  publictestnet.LiveEvidenceWindow `json:"evidence"`
		Readiness publictestnet.EvidenceReadiness  `json:"readiness"`
	}{window, readiness}
	raw, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		fatal(err)
	}
	if *outPath == "" {
		fmt.Println(string(raw))
	} else if err := os.WriteFile(*outPath, append(raw, '\n'), 0o600); err != nil {
		fatal(err)
	}
	if *gate && !readiness.Ready {
		fatal(fmt.Errorf("live evidence gate not ready: %s", strings.Join(readiness.Reasons, "; ")))
	}
}

func loadOperators(path string) ([]publictestnet.OperatorDescriptor, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var operators []publictestnet.OperatorDescriptor
	if err := json.Unmarshal(raw, &operators); err != nil {
		return nil, fmt.Errorf("decode operators: %w", err)
	}
	for i := range operators {
		operators[i].ID = strings.TrimSpace(operators[i].ID)
		operators[i].URL = strings.TrimRight(strings.TrimSpace(operators[i].URL), "/")
	}
	return operators, nil
}

func collect(client *http.Client, sequence int, operator publictestnet.OperatorDescriptor) publictestnet.LiveEvidenceSample {
	sample := publictestnet.LiveEvidenceSample{Sequence: sequence, CollectedAt: time.Now().UTC(), OperatorID: operator.ID}
	started := time.Now()
	status, err := getMap(client, operator.URL+"/v1/public/status")
	sample.StatusLatencyMillis = time.Since(started).Milliseconds()
	if err != nil {
		sample.Error = "status: " + err.Error()
		return sample
	}
	sample.Status = status
	started = time.Now()
	checkpoint, err := getCheckpoint(client, operator.URL+"/v1/public/checkpoint")
	sample.CheckpointLatencyMillis = time.Since(started).Milliseconds()
	if err != nil {
		sample.Error = "checkpoint: " + err.Error()
		return sample
	}
	sample.Checkpoint = &checkpoint
	sample.Available = true
	return sample
}

func getMap(client *http.Client, url string) (map[string]any, error) {
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
	var value publictestnet.Checkpoint
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&value); err != nil {
		return value, err
	}
	return value, nil
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "pop-live-evidence:", err)
	os.Exit(1)
}
