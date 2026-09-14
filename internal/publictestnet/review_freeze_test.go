package publictestnet

import (
	"encoding/json"
	"testing"
	"time"
)

func TestBuildReviewFreeze(t *testing.T) {
	start := time.Unix(1_800_000_000, 0).UTC()
	checkpoint := liveEvidenceCheckpoint("network", "genesis", "policy", 7, "block", "state")
	window := LiveEvidenceWindow{Version: LiveEvidenceVersion, RepositorySHA: "0123456789012345678901234567890123456789", NetworkID: "network", GenesisHash: "genesis", CarrotPolicyHash: "policy", NetworkMapHash: "map", StartedAt: start, EndedAt: start.Add(time.Hour), Operators: []OperatorDescriptor{{ID: "a", URL: "https://a.example"}, {ID: "b", URL: "https://b.example"}}, Samples: []LiveEvidenceSample{{Sequence: 1, CollectedAt: start.Add(time.Minute), OperatorID: "a", Available: true, Checkpoint: &checkpoint}, {Sequence: 1, CollectedAt: start.Add(time.Minute), OperatorID: "b", Available: true, Checkpoint: &checkpoint}}}
	if err := window.Finalize(); err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(window)
	digest := "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	manifest, err := BuildReviewFreeze(window.RepositorySHA, []NamedEvidenceWindow{{Name: "window.json", Raw: raw, Window: window}}, start.Add(2*time.Hour), digest)
	if err != nil {
		t.Fatal(err)
	}
	if err := manifest.Validate(); err != nil {
		t.Fatal(err)
	}
	if manifest.Hash == "" || len(manifest.Evidence) != 1 || manifest.NetworkMapHashes[0] != "map" || len(manifest.ImageDigests) != 1 || manifest.ImageDigests[0] != digest {
		t.Fatalf("manifest=%+v", manifest)
	}
}

func TestReviewFreezeRejectsDifferentRepositorySHA(t *testing.T) {
	start := time.Unix(1_800_000_000, 0).UTC()
	checkpoint := liveEvidenceCheckpoint("network", "genesis", "policy", 7, "block", "state")
	window := LiveEvidenceWindow{Version: LiveEvidenceVersion, RepositorySHA: "other", NetworkID: "network", GenesisHash: "genesis", CarrotPolicyHash: "policy", StartedAt: start, EndedAt: start.Add(time.Hour), Operators: []OperatorDescriptor{{ID: "a", URL: "https://a.example"}, {ID: "b", URL: "https://b.example"}}, Samples: []LiveEvidenceSample{{Sequence: 1, CollectedAt: start.Add(time.Minute), OperatorID: "a", Available: true, Checkpoint: &checkpoint}, {Sequence: 1, CollectedAt: start.Add(time.Minute), OperatorID: "b", Available: true, Checkpoint: &checkpoint}}}
	if err := window.Finalize(); err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(window)
	if _, err := BuildReviewFreeze("wanted", []NamedEvidenceWindow{{Name: "window.json", Raw: raw, Window: window}}, start.Add(2*time.Hour)); err == nil {
		t.Fatal("expected repository SHA mismatch")
	}
}

func TestReviewFreezeRejectsInvalidImageDigest(t *testing.T) {
	start := time.Unix(1_800_000_000, 0).UTC()
	checkpoint := liveEvidenceCheckpoint("network", "genesis", "policy", 7, "block", "state")
	window := LiveEvidenceWindow{Version: LiveEvidenceVersion, RepositorySHA: "repo", NetworkID: "network", GenesisHash: "genesis", CarrotPolicyHash: "policy", StartedAt: start, EndedAt: start.Add(time.Hour), Operators: []OperatorDescriptor{{ID: "a", URL: "https://a.example"}, {ID: "b", URL: "https://b.example"}}, Samples: []LiveEvidenceSample{{Sequence: 1, CollectedAt: start.Add(time.Minute), OperatorID: "a", Available: true, Checkpoint: &checkpoint}, {Sequence: 1, CollectedAt: start.Add(time.Minute), OperatorID: "b", Available: true, Checkpoint: &checkpoint}}}
	if err := window.Finalize(); err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(window)
	if _, err := BuildReviewFreeze("repo", []NamedEvidenceWindow{{Name: "window.json", Raw: raw, Window: window}}, start.Add(time.Hour), "sha256:not-a-digest"); err == nil {
		t.Fatal("expected invalid digest")
	}
}
