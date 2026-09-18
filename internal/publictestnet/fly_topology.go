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

const FlyTopologyEvidenceVersion = 1

var ErrInvalidFlyTopologyEvidence = errors.New("invalid Fly topology evidence")

var flyEvidenceRegions = []string{"iad", "ord", "dfw", "lax"}

type NamedFlyMachineList struct {
	OperatorID     string
	AppName        string
	ExpectedRegion string
	Raw            []byte
}

type flyMachineListEntry struct {
	ID       string `json:"id"`
	State    string `json:"state"`
	Region   string `json:"region"`
	ImageRef struct {
		Digest string `json:"digest"`
	} `json:"image_ref"`
}

type FlyTopologyNodeEvidence struct {
	OperatorID            string `json:"operatorId"`
	AppName               string `json:"appName"`
	Region                string `json:"region"`
	MachineID             string `json:"machineId"`
	State                 string `json:"state"`
	ImageDigest           string `json:"imageDigest"`
	MachineListSHA256     string `json:"machineListSha256"`
}

type FlyTopologyEvidence struct {
	Version       int                       `json:"version"`
	RepositorySHA string                    `json:"repositorySha"`
	CollectedAt   time.Time                 `json:"collectedAt"`
	Nodes         []FlyTopologyNodeEvidence `json:"nodes"`
	ImageDigests  []string                  `json:"imageDigests"`
	Hash          string                    `json:"hash"`
}

func BuildFlyTopologyEvidence(repositorySHA string, collectedAt time.Time, inputs []NamedFlyMachineList) (FlyTopologyEvidence, error) {
	repositorySHA = strings.TrimSpace(repositorySHA)
	if repositorySHA == "" || collectedAt.IsZero() || len(inputs) != len(flyEvidenceRegions) {
		return FlyTopologyEvidence{}, ErrInvalidFlyTopologyEvidence
	}
	evidence := FlyTopologyEvidence{
		Version:       FlyTopologyEvidenceVersion,
		RepositorySHA: repositorySHA,
		CollectedAt:   collectedAt.UTC(),
	}
	operators := map[string]struct{}{}
	apps := map[string]struct{}{}
	regions := map[string]struct{}{}
	digests := map[string]struct{}{}
	for _, input := range inputs {
		operatorID := strings.TrimSpace(input.OperatorID)
		appName := strings.TrimSpace(input.AppName)
		region := strings.ToLower(strings.TrimSpace(input.ExpectedRegion))
		if operatorID == "" || appName == "" || !isFlyEvidenceRegion(region) || len(input.Raw) == 0 {
			return FlyTopologyEvidence{}, ErrInvalidFlyTopologyEvidence
		}
		if _, exists := operators[operatorID]; exists {
			return FlyTopologyEvidence{}, fmt.Errorf("%w: duplicate operator %q", ErrInvalidFlyTopologyEvidence, operatorID)
		}
		if _, exists := apps[appName]; exists {
			return FlyTopologyEvidence{}, fmt.Errorf("%w: duplicate Fly app %q", ErrInvalidFlyTopologyEvidence, appName)
		}
		if _, exists := regions[region]; exists {
			return FlyTopologyEvidence{}, fmt.Errorf("%w: duplicate region %q", ErrInvalidFlyTopologyEvidence, region)
		}
		operators[operatorID] = struct{}{}
		apps[appName] = struct{}{}
		regions[region] = struct{}{}

		var machines []flyMachineListEntry
		if err := json.Unmarshal(input.Raw, &machines); err != nil {
			return FlyTopologyEvidence{}, fmt.Errorf("%w: decode %s machine list: %v", ErrInvalidFlyTopologyEvidence, operatorID, err)
		}
		var started []flyMachineListEntry
		for _, machine := range machines {
			if strings.EqualFold(strings.TrimSpace(machine.State), "started") {
				started = append(started, machine)
			}
		}
		if len(started) != 1 {
			return FlyTopologyEvidence{}, fmt.Errorf("%w: %s has %d started machines, want 1", ErrInvalidFlyTopologyEvidence, operatorID, len(started))
		}
		machine := started[0]
		machineRegion := strings.ToLower(strings.TrimSpace(machine.Region))
		digest := strings.ToLower(strings.TrimSpace(machine.ImageRef.Digest))
		if strings.TrimSpace(machine.ID) == "" || machineRegion != region {
			return FlyTopologyEvidence{}, fmt.Errorf("%w: %s started machine region=%q want=%q", ErrInvalidFlyTopologyEvidence, operatorID, machineRegion, region)
		}
		if !validSHA256Digest(digest) {
			return FlyTopologyEvidence{}, fmt.Errorf("%w: %s invalid image digest %q", ErrInvalidFlyTopologyEvidence, operatorID, digest)
		}
		rawSum := sha256.Sum256(input.Raw)
		evidence.Nodes = append(evidence.Nodes, FlyTopologyNodeEvidence{
			OperatorID:        operatorID,
			AppName:           appName,
			Region:            region,
			MachineID:         strings.TrimSpace(machine.ID),
			State:             "started",
			ImageDigest:       digest,
			MachineListSHA256: hex.EncodeToString(rawSum[:]),
		})
		digests[digest] = struct{}{}
	}
	for _, region := range flyEvidenceRegions {
		if _, ok := regions[region]; !ok {
			return FlyTopologyEvidence{}, fmt.Errorf("%w: missing region %q", ErrInvalidFlyTopologyEvidence, region)
		}
	}
	sort.Slice(evidence.Nodes, func(i, j int) bool {
		return evidence.Nodes[i].OperatorID < evidence.Nodes[j].OperatorID
	})
	for digest := range digests {
		evidence.ImageDigests = append(evidence.ImageDigests, digest)
	}
	sort.Strings(evidence.ImageDigests)
	evidence.Hash = flyTopologyEvidenceHash(evidence)
	if err := evidence.Validate(); err != nil {
		return FlyTopologyEvidence{}, err
	}
	return evidence, nil
}

func (e FlyTopologyEvidence) Validate() error {
	if e.Version != FlyTopologyEvidenceVersion || strings.TrimSpace(e.RepositorySHA) == "" || e.CollectedAt.IsZero() || len(e.Nodes) != len(flyEvidenceRegions) || len(e.ImageDigests) == 0 {
		return ErrInvalidFlyTopologyEvidence
	}
	operators := map[string]struct{}{}
	apps := map[string]struct{}{}
	regions := map[string]struct{}{}
	digests := map[string]struct{}{}
	for _, node := range e.Nodes {
		if strings.TrimSpace(node.OperatorID) == "" || strings.TrimSpace(node.AppName) == "" || strings.TrimSpace(node.MachineID) == "" || node.State != "started" {
			return ErrInvalidFlyTopologyEvidence
		}
		if _, ok := operators[node.OperatorID]; ok {
			return ErrInvalidFlyTopologyEvidence
		}
		if _, ok := apps[node.AppName]; ok {
			return ErrInvalidFlyTopologyEvidence
		}
		if _, ok := regions[node.Region]; ok || !isFlyEvidenceRegion(node.Region) {
			return ErrInvalidFlyTopologyEvidence
		}
		if !validSHA256Digest(node.ImageDigest) || !validHexSHA256(node.MachineListSHA256) {
			return ErrInvalidFlyTopologyEvidence
		}
		operators[node.OperatorID] = struct{}{}
		apps[node.AppName] = struct{}{}
		regions[node.Region] = struct{}{}
		digests[node.ImageDigest] = struct{}{}
	}
	for _, region := range flyEvidenceRegions {
		if _, ok := regions[region]; !ok {
			return ErrInvalidFlyTopologyEvidence
		}
	}
	expectedDigests := make([]string, 0, len(digests))
	for digest := range digests {
		expectedDigests = append(expectedDigests, digest)
	}
	sort.Strings(expectedDigests)
	if len(expectedDigests) != len(e.ImageDigests) {
		return ErrInvalidFlyTopologyEvidence
	}
	for i, digest := range expectedDigests {
		if e.ImageDigests[i] != digest {
			return ErrInvalidFlyTopologyEvidence
		}
	}
	if e.Hash == "" || e.Hash != flyTopologyEvidenceHash(e) {
		return fmt.Errorf("%w: topology hash mismatch", ErrInvalidFlyTopologyEvidence)
	}
	return nil
}

func isFlyEvidenceRegion(region string) bool {
	for _, expected := range flyEvidenceRegions {
		if region == expected {
			return true
		}
	}
	return false
}

func validHexSHA256(value string) bool {
	if len(value) != 64 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func flyTopologyEvidenceHash(evidence FlyTopologyEvidence) string {
	evidence.Hash = ""
	raw, _ := json.Marshal(evidence)
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}
