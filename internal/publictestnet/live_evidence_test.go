package publictestnet

import (
	"testing"
	"time"
)

func liveEvidenceCheckpoint(network, genesis, policy string, height uint64, finalized, state string) Checkpoint {
	checkpoint := Checkpoint{Version: CheckpointVersion, NetworkID: network, GenesisHash: genesis, Height: height, FinalizedHash: finalized, StateRoot: state, ProtocolVersion: 3, ValidatorSetHash: "validators", CarrotPolicyHash: policy}
	checkpoint.Hash = checkpointHash(checkpoint)
	return checkpoint
}

func TestLiveEvidenceFinalizesAndValidates(t *testing.T) {
	start := time.Unix(1_800_000_000, 0).UTC()
	checkpoint := liveEvidenceCheckpoint("network", "genesis", "policy", 42, "block", "state")
	window := LiveEvidenceWindow{Version: LiveEvidenceVersion, RepositorySHA: "0123456789012345678901234567890123456789", NetworkID: "network", GenesisHash: "genesis", CarrotPolicyHash: "policy", NetworkMapHash: "map", StartedAt: start, EndedAt: start.Add(time.Hour), Operators: []OperatorDescriptor{{ID: "a", URL: "https://a.example", Provider: "provider-a", Region: "us-east"}, {ID: "b", URL: "https://b.example", Provider: "provider-b", Region: "us-west"}}, Samples: []LiveEvidenceSample{{Sequence: 1, CollectedAt: start.Add(time.Minute), OperatorID: "a", Available: true, Checkpoint: &checkpoint}, {Sequence: 1, CollectedAt: start.Add(2 * time.Minute), OperatorID: "b", Available: true, Checkpoint: &checkpoint}}}
	if err := window.Finalize(); err != nil {
		t.Fatal(err)
	}
	if err := window.Validate(); err != nil {
		t.Fatal(err)
	}
	readiness := window.EvaluateReadiness(EvidenceReadinessPolicy{MinOperators: 2, MinProviders: 2, MinRegions: 2, MinSamplesPerOperator: 1, MinWindow: time.Hour})
	if !readiness.Ready {
		t.Fatalf("readiness=%+v", readiness)
	}
}

func TestLiveEvidencePreservesSameHeightDisagreement(t *testing.T) {
	start := time.Unix(1_800_000_000, 0).UTC()
	a := liveEvidenceCheckpoint("network", "genesis", "policy", 42, "block-a", "state-a")
	b := liveEvidenceCheckpoint("network", "genesis", "policy", 42, "block-b", "state-b")
	window := LiveEvidenceWindow{Version: LiveEvidenceVersion, RepositorySHA: "sha", NetworkID: "network", GenesisHash: "genesis", CarrotPolicyHash: "policy", StartedAt: start, EndedAt: start.Add(time.Hour), Operators: []OperatorDescriptor{{ID: "a", URL: "https://a.example", Provider: "one"}, {ID: "b", URL: "https://b.example", Provider: "two"}}, Samples: []LiveEvidenceSample{{Sequence: 1, CollectedAt: start.Add(time.Minute), OperatorID: "a", Available: true, Checkpoint: &a}, {Sequence: 1, CollectedAt: start.Add(time.Minute), OperatorID: "b", Available: true, Checkpoint: &b}}}
	if err := window.Finalize(); err != nil {
		t.Fatal(err)
	}
	if err := window.Validate(); err != nil {
		t.Fatal(err)
	}
	if len(window.Incidents) != 1 || window.Incidents[0].Kind != IncidentCheckpointMismatch || window.Incidents[0].Height != 42 {
		t.Fatalf("incidents=%+v", window.Incidents)
	}
	if readiness := window.EvaluateReadiness(EvidenceReadinessPolicy{MinOperators: 2, MinProviders: 2}); readiness.Ready {
		t.Fatalf("unexpected readiness=%+v", readiness)
	}
}

func TestLiveEvidenceDetectsCrossSampleSameHeightDisagreement(t *testing.T) {
	start := time.Unix(1_800_000_000, 0).UTC()
	a := liveEvidenceCheckpoint("network", "genesis", "policy", 42, "block-a", "state-a")
	b := liveEvidenceCheckpoint("network", "genesis", "policy", 42, "block-b", "state-b")
	window := LiveEvidenceWindow{Version: LiveEvidenceVersion, RepositorySHA: "sha", NetworkID: "network", GenesisHash: "genesis", CarrotPolicyHash: "policy", StartedAt: start, EndedAt: start.Add(time.Hour), Operators: []OperatorDescriptor{{ID: "a", URL: "https://a.example", Provider: "one"}, {ID: "b", URL: "https://b.example", Provider: "two"}}, Samples: []LiveEvidenceSample{{Sequence: 1, CollectedAt: start.Add(time.Minute), OperatorID: "a", Available: true, Checkpoint: &a}, {Sequence: 2, CollectedAt: start.Add(5 * time.Minute), OperatorID: "b", Available: true, Checkpoint: &b}}}
	if err := window.Finalize(); err != nil {
		t.Fatal(err)
	}
	if len(window.Incidents) != 1 || window.Incidents[0].Height != 42 {
		t.Fatalf("incidents=%+v", window.Incidents)
	}
}

func TestLiveEvidenceAllowsHeightLag(t *testing.T) {
	start := time.Unix(1_800_000_000, 0).UTC()
	a := liveEvidenceCheckpoint("network", "genesis", "policy", 42, "block-a", "state-a")
	b := liveEvidenceCheckpoint("network", "genesis", "policy", 41, "block-b", "state-b")
	window := LiveEvidenceWindow{Version: LiveEvidenceVersion, RepositorySHA: "sha", NetworkID: "network", GenesisHash: "genesis", CarrotPolicyHash: "policy", StartedAt: start, EndedAt: start.Add(time.Hour), Operators: []OperatorDescriptor{{ID: "a", URL: "https://a.example"}, {ID: "b", URL: "https://b.example"}}, Samples: []LiveEvidenceSample{{Sequence: 1, CollectedAt: start.Add(time.Minute), OperatorID: "a", Available: true, Checkpoint: &a}, {Sequence: 1, CollectedAt: start.Add(time.Minute), OperatorID: "b", Available: true, Checkpoint: &b}}}
	if err := window.Finalize(); err != nil {
		t.Fatal(err)
	}
	if len(window.Incidents) != 0 {
		t.Fatalf("incidents=%+v", window.Incidents)
	}
}

func TestLiveEvidenceReadinessCountsOnlyAvailableSamples(t *testing.T) {
	start := time.Unix(1_800_000_000, 0).UTC()
	checkpoint := liveEvidenceCheckpoint("network", "genesis", "policy", 42, "block", "state")
	window := LiveEvidenceWindow{Version: LiveEvidenceVersion, RepositorySHA: "sha", NetworkID: "network", GenesisHash: "genesis", CarrotPolicyHash: "policy", StartedAt: start, EndedAt: start.Add(time.Hour), Operators: []OperatorDescriptor{{ID: "a", URL: "https://a.example", Provider: "one"}, {ID: "b", URL: "https://b.example", Provider: "two"}}, Samples: []LiveEvidenceSample{{Sequence: 1, CollectedAt: start.Add(time.Minute), OperatorID: "a", Available: true, Checkpoint: &checkpoint}, {Sequence: 1, CollectedAt: start.Add(time.Minute), OperatorID: "b", Available: false, Error: "offline"}}}
	if err := window.Finalize(); err != nil {
		t.Fatal(err)
	}
	if readiness := window.EvaluateReadiness(EvidenceReadinessPolicy{MinOperators: 2, MinProviders: 2, MinSamplesPerOperator: 1}); readiness.Ready || readiness.MinSamplesObserved != 0 {
		t.Fatalf("readiness=%+v", readiness)
	}
}
