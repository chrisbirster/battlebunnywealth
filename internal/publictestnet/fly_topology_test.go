package publictestnet

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func flyMachineListJSON(t *testing.T, id, region, state, digest string) []byte {
	t.Helper()
	raw, err := json.Marshal([]map[string]any{{
		"id":     id,
		"region": region,
		"state":  state,
		"image_ref": map[string]any{
			"digest": digest,
		},
	}})
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func flyTopologyInputs(t *testing.T, digest string) []NamedFlyMachineList {
	t.Helper()
	inputs := make([]NamedFlyMachineList, 0, len(flyEvidenceRegions))
	for _, region := range flyEvidenceRegions {
		inputs = append(inputs, NamedFlyMachineList{
			OperatorID:     region,
			AppName:        "bbwealth-pop-" + region,
			ExpectedRegion: region,
			Raw:            flyMachineListJSON(t, "machine-"+region, region, "started", digest),
		})
	}
	return inputs
}

func TestFlyTopologyEvidenceBuildsAndValidates(t *testing.T) {
	digest := "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	evidence, err := BuildFlyTopologyEvidence(
		"35c3340ed7b8168066edd26aaf67e678d1e9ad5e",
		time.Unix(1_800_000_000, 0).UTC(),
		flyTopologyInputs(t, digest),
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := evidence.Validate(); err != nil {
		t.Fatal(err)
	}
	if len(evidence.Nodes) != 4 || len(evidence.ImageDigests) != 1 || evidence.ImageDigests[0] != digest || evidence.Hash == "" {
		t.Fatalf("evidence=%+v", evidence)
	}
	for _, node := range evidence.Nodes {
		if node.MachineListSHA256 == "" || node.State != "started" {
			t.Fatalf("node=%+v", node)
		}
	}
}

func TestFlyTopologyEvidenceAllowsMultipleImmutableBuildDigests(t *testing.T) {
	first := "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	second := "sha256:abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"
	inputs := flyTopologyInputs(t, first)
	inputs[3].Raw = flyMachineListJSON(t, "machine-lax", "lax", "started", second)
	evidence, err := BuildFlyTopologyEvidence("repo", time.Unix(1_800_000_000, 0).UTC(), inputs)
	if err != nil {
		t.Fatal(err)
	}
	if len(evidence.ImageDigests) != 2 {
		t.Fatalf("imageDigests=%v", evidence.ImageDigests)
	}
}

func TestFlyTopologyEvidenceRejectsRegionMismatch(t *testing.T) {
	digest := "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	inputs := flyTopologyInputs(t, digest)
	inputs[0].Raw = flyMachineListJSON(t, "machine-iad", "ord", "started", digest)
	if _, err := BuildFlyTopologyEvidence("repo", time.Unix(1_800_000_000, 0).UTC(), inputs); err == nil {
		t.Fatal("expected region mismatch")
	}
}

func TestFlyTopologyEvidenceRejectsMultipleStartedMachines(t *testing.T) {
	digest := "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	inputs := flyTopologyInputs(t, digest)
	var machines []map[string]any
	if err := json.Unmarshal(inputs[0].Raw, &machines); err != nil {
		t.Fatal(err)
	}
	machines = append(machines, map[string]any{
		"id": "machine-iad-2", "region": "iad", "state": "started",
		"image_ref": map[string]any{"digest": digest},
	})
	raw, err := json.Marshal(machines)
	if err != nil {
		t.Fatal(err)
	}
	inputs[0].Raw = raw
	if _, err := BuildFlyTopologyEvidence("repo", time.Unix(1_800_000_000, 0).UTC(), inputs); err == nil {
		t.Fatal("expected multiple-started-machine rejection")
	}
}

func TestFlyTopologyEvidenceRejectsInvalidDigest(t *testing.T) {
	inputs := flyTopologyInputs(t, "sha256:not-a-digest")
	_, err := BuildFlyTopologyEvidence("repo", time.Unix(1_800_000_000, 0).UTC(), inputs)
	if err == nil {
		t.Fatal("expected invalid digest")
	}
	if got := fmt.Sprint(err); got == "" {
		t.Fatal("expected useful error")
	}
}
