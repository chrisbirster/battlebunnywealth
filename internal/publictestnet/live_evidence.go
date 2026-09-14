package publictestnet

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

const LiveEvidenceVersion = 1

var (
	ErrInvalidLiveEvidence = errors.New("invalid live public-testnet evidence")
	ErrLiveConsensusMismatch = errors.New("live public-testnet consensus mismatch")
)

type OperatorDescriptor struct {
	ID       string `json:"id"`
	URL      string `json:"url"`
	Provider string `json:"provider,omitempty"`
	Region   string `json:"region,omitempty"`
	ASN      int    `json:"asn,omitempty"`
}

type LiveEvidenceSample struct {
	Sequence                int            `json:"sequence"`
	CollectedAt             time.Time      `json:"collectedAt"`
	OperatorID              string         `json:"operatorId"`
	Available               bool           `json:"available"`
	StatusLatencyMillis     int64          `json:"statusLatencyMillis,omitempty"`
	CheckpointLatencyMillis int64          `json:"checkpointLatencyMillis,omitempty"`
	Status                  map[string]any `json:"status,omitempty"`
	Checkpoint              *Checkpoint    `json:"checkpoint,omitempty"`
	Error                   string         `json:"error,omitempty"`
}

type EvidenceIncident struct {
	Kind       string    `json:"kind"`
	ObservedAt time.Time `json:"observedAt"`
	Sequence   int       `json:"sequence,omitempty"`
	Details    string    `json:"details"`
}

type LiveEvidenceWindow struct {
	Version          int                  `json:"version"`
	RepositorySHA    string               `json:"repositorySha"`
	NetworkID        string               `json:"networkId"`
	GenesisHash      string               `json:"genesisHash"`
	CarrotPolicyHash string               `json:"carrotPolicyHash"`
	NetworkMapHash   string               `json:"networkMapHash,omitempty"`
	StartedAt        time.Time            `json:"startedAt"`
	EndedAt          time.Time            `json:"endedAt"`
	Operators        []OperatorDescriptor `json:"operators"`
	Samples          []LiveEvidenceSample `json:"samples"`
	Incidents        []EvidenceIncident   `json:"incidents,omitempty"`
	Hash             string               `json:"hash"`
}

type EvidenceReadinessPolicy struct {
	MinOperators          int           `json:"minOperators"`
	MinProviders          int           `json:"minProviders"`
	MinRegions            int           `json:"minRegions"`
	MinSamplesPerOperator int           `json:"minSamplesPerOperator"`
	MinWindow             time.Duration `json:"-"`
}

type EvidenceReadiness struct {
	Ready              bool     `json:"ready"`
	Operators          int      `json:"operators"`
	Providers          int      `json:"providers"`
	Regions            int      `json:"regions"`
	MinSamplesObserved int      `json:"minSamplesObserved"`
	Reasons            []string `json:"reasons,omitempty"`
}

func (w *LiveEvidenceWindow) Finalize() error {
	if w.Version == 0 {
		w.Version = LiveEvidenceVersion
	}
	if err := w.validate(false); err != nil {
		return err
	}
	sort.Slice(w.Operators, func(i, j int) bool { return w.Operators[i].ID < w.Operators[j].ID })
	sort.Slice(w.Samples, func(i, j int) bool {
		if w.Samples[i].Sequence == w.Samples[j].Sequence {
			return w.Samples[i].OperatorID < w.Samples[j].OperatorID
		}
		return w.Samples[i].Sequence < w.Samples[j].Sequence
	})
	sort.Slice(w.Incidents, func(i, j int) bool {
		if w.Incidents[i].Sequence == w.Incidents[j].Sequence {
			return w.Incidents[i].ObservedAt.Before(w.Incidents[j].ObservedAt)
		}
		return w.Incidents[i].Sequence < w.Incidents[j].Sequence
	})
	w.Hash = liveEvidenceHash(*w)
	return nil
}

func (w LiveEvidenceWindow) Validate() error {
	if err := w.validate(true); err != nil {
		return err
	}
	if w.Hash == "" || w.Hash != liveEvidenceHash(w) {
		return fmt.Errorf("%w: evidence hash mismatch", ErrInvalidLiveEvidence)
	}
	return nil
}

func (w LiveEvidenceWindow) validate(requireHash bool) error {
	if w.Version != LiveEvidenceVersion {
		return fmt.Errorf("%w: unsupported version %d", ErrInvalidLiveEvidence, w.Version)
	}
	if strings.TrimSpace(w.RepositorySHA) == "" || strings.TrimSpace(w.NetworkID) == "" || strings.TrimSpace(w.GenesisHash) == "" || strings.TrimSpace(w.CarrotPolicyHash) == "" {
		return fmt.Errorf("%w: missing repository/network/genesis/policy commitment", ErrInvalidLiveEvidence)
	}
	if w.StartedAt.IsZero() || w.EndedAt.IsZero() || w.EndedAt.Before(w.StartedAt) {
		return fmt.Errorf("%w: invalid evidence window", ErrInvalidLiveEvidence)
	}
	if len(w.Operators) < 2 {
		return fmt.Errorf("%w: at least two operators required", ErrInvalidLiveEvidence)
	}
	operators := map[string]OperatorDescriptor{}
	urls := map[string]struct{}{}
	for _, operator := range w.Operators {
		operator.ID = strings.TrimSpace(operator.ID)
		operator.URL = strings.TrimRight(strings.TrimSpace(operator.URL), "/")
		if operator.ID == "" || operator.URL == "" {
			return fmt.Errorf("%w: operator id and url required", ErrInvalidLiveEvidence)
		}
		if _, ok := operators[operator.ID]; ok {
			return fmt.Errorf("%w: duplicate operator %q", ErrInvalidLiveEvidence, operator.ID)
		}
		if _, ok := urls[operator.URL]; ok {
			return fmt.Errorf("%w: duplicate operator url %q", ErrInvalidLiveEvidence, operator.URL)
		}
		operators[operator.ID] = operator
		urls[operator.URL] = struct{}{}
	}
	if len(w.Samples) == 0 {
		return fmt.Errorf("%w: no samples", ErrInvalidLiveEvidence)
	}
	seenSamples := map[string]struct{}{}
	bySequence := map[int][]CheckpointObservation{}
	for _, sample := range w.Samples {
		if sample.Sequence <= 0 || sample.OperatorID == "" || sample.CollectedAt.IsZero() {
			return fmt.Errorf("%w: incomplete sample", ErrInvalidLiveEvidence)
		}
		if _, ok := operators[sample.OperatorID]; !ok {
			return fmt.Errorf("%w: sample references unknown operator %q", ErrInvalidLiveEvidence, sample.OperatorID)
		}
		if sample.CollectedAt.Before(w.StartedAt) || sample.CollectedAt.After(w.EndedAt) {
			return fmt.Errorf("%w: sample outside evidence window", ErrInvalidLiveEvidence)
		}
		key := fmt.Sprintf("%d/%s", sample.Sequence, sample.OperatorID)
		if _, ok := seenSamples[key]; ok {
			return fmt.Errorf("%w: duplicate sample %s", ErrInvalidLiveEvidence, key)
		}
		seenSamples[key] = struct{}{}
		if sample.Available {
			if sample.Checkpoint == nil {
				return fmt.Errorf("%w: available sample missing checkpoint", ErrInvalidLiveEvidence)
			}
			checkpoint := *sample.Checkpoint
			if checkpoint.Version != CheckpointVersion || checkpoint.Hash != checkpointHash(checkpoint) || checkpoint.NetworkID != w.NetworkID || checkpoint.GenesisHash != w.GenesisHash || checkpoint.CarrotPolicyHash != w.CarrotPolicyHash {
				return fmt.Errorf("%w: invalid checkpoint from %s", ErrInvalidLiveEvidence, sample.OperatorID)
			}
			bySequence[sample.Sequence] = append(bySequence[sample.Sequence], CheckpointObservation{Source: sample.OperatorID, Checkpoint: checkpoint})
		}
	}
	for sequence, observations := range bySequence {
		byHeight := map[uint64][]CheckpointObservation{}
		for _, observation := range observations {
			byHeight[observation.Checkpoint.Height] = append(byHeight[observation.Checkpoint.Height], observation)
		}
		for height, sameHeight := range byHeight {
			if len(sameHeight) < 2 {
				continue
			}
			if _, err := CompareCheckpointSet(sameHeight); err != nil {
				return fmt.Errorf("%w: sequence %d height %d: %v", ErrLiveConsensusMismatch, sequence, height, err)
			}
		}
	}
	if requireHash && w.Hash == "" {
		return fmt.Errorf("%w: missing evidence hash", ErrInvalidLiveEvidence)
	}
	return nil
}

func (w LiveEvidenceWindow) EvaluateReadiness(policy EvidenceReadinessPolicy) EvidenceReadiness {
	if policy.MinOperators <= 0 { policy.MinOperators = 2 }
	if policy.MinProviders <= 0 { policy.MinProviders = 2 }
	if policy.MinRegions <= 0 { policy.MinRegions = 1 }
	if policy.MinSamplesPerOperator <= 0 { policy.MinSamplesPerOperator = 1 }

	providers := map[string]struct{}{}
	regions := map[string]struct{}{}
	counts := map[string]int{}
	for _, operator := range w.Operators {
		if value := strings.ToLower(strings.TrimSpace(operator.Provider)); value != "" { providers[value] = struct{}{} }
		if value := strings.ToLower(strings.TrimSpace(operator.Region)); value != "" { regions[value] = struct{}{} }
	}
	for _, sample := range w.Samples { counts[sample.OperatorID]++ }
	minSamples := 0
	for _, operator := range w.Operators {
		count := counts[operator.ID]
		if minSamples == 0 || count < minSamples { minSamples = count }
	}
	result := EvidenceReadiness{Operators: len(w.Operators), Providers: len(providers), Regions: len(regions), MinSamplesObserved: minSamples}
	if result.Operators < policy.MinOperators { result.Reasons = append(result.Reasons, fmt.Sprintf("operators=%d want>=%d", result.Operators, policy.MinOperators)) }
	if result.Providers < policy.MinProviders { result.Reasons = append(result.Reasons, fmt.Sprintf("providers=%d want>=%d", result.Providers, policy.MinProviders)) }
	if result.Regions < policy.MinRegions { result.Reasons = append(result.Reasons, fmt.Sprintf("regions=%d want>=%d", result.Regions, policy.MinRegions)) }
	if result.MinSamplesObserved < policy.MinSamplesPerOperator { result.Reasons = append(result.Reasons, fmt.Sprintf("samplesPerOperator=%d want>=%d", result.MinSamplesObserved, policy.MinSamplesPerOperator)) }
	if policy.MinWindow > 0 && w.EndedAt.Sub(w.StartedAt) < policy.MinWindow { result.Reasons = append(result.Reasons, fmt.Sprintf("window=%s want>=%s", w.EndedAt.Sub(w.StartedAt), policy.MinWindow)) }
	if err := w.Validate(); err != nil { result.Reasons = append(result.Reasons, err.Error()) }
	result.Ready = len(result.Reasons) == 0
	return result
}

func liveEvidenceHash(window LiveEvidenceWindow) string {
	window.Hash = ""
	raw, _ := json.Marshal(window)
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}
