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

const ReviewFreezeVersion = 1

var ErrInvalidReviewFreeze = errors.New("invalid independent-review freeze manifest")

type EvidenceArtifactDigest struct {
	Name         string `json:"name"`
	SHA256       string `json:"sha256"`
	EvidenceHash string `json:"evidenceHash"`
}

type ReviewFreezeManifest struct {
	Version          int                      `json:"version"`
	RepositorySHA    string                   `json:"repositorySha"`
	ImageDigests     []string                 `json:"imageDigests"`
	NetworkID        string                   `json:"networkId"`
	GenesisHash      string                   `json:"genesisHash"`
	CarrotPolicyHash string                   `json:"carrotPolicyHash"`
	NetworkMapHashes []string                 `json:"networkMapHashes,omitempty"`
	CreatedAt        time.Time                `json:"createdAt"`
	Evidence         []EvidenceArtifactDigest `json:"evidence"`
	Hash             string                   `json:"hash"`
}

type NamedEvidenceWindow struct {
	Name   string
	Raw    []byte
	Window LiveEvidenceWindow
}

func BuildReviewFreeze(repositorySHA string, artifacts []NamedEvidenceWindow, createdAt time.Time, imageDigests ...string) (ReviewFreezeManifest, error) {
	if strings.TrimSpace(repositorySHA) == "" || len(artifacts) == 0 || createdAt.IsZero() {
		return ReviewFreezeManifest{}, ErrInvalidReviewFreeze
	}
	manifest := ReviewFreezeManifest{Version: ReviewFreezeVersion, RepositorySHA: repositorySHA, CreatedAt: createdAt.UTC()}
	imageSet := map[string]struct{}{}
	for _, digest := range imageDigests {
		digest = strings.ToLower(strings.TrimSpace(digest))
		if digest == "" {
			continue
		}
		if !validSHA256Digest(digest) {
			return ReviewFreezeManifest{}, fmt.Errorf("%w: invalid image digest %q", ErrInvalidReviewFreeze, digest)
		}
		imageSet[digest] = struct{}{}
	}
	if len(imageSet) == 0 {
		return ReviewFreezeManifest{}, fmt.Errorf("%w: at least one immutable image digest is required", ErrInvalidReviewFreeze)
	}
	for digest := range imageSet {
		manifest.ImageDigests = append(manifest.ImageDigests, digest)
	}
	sort.Strings(manifest.ImageDigests)
	maps := map[string]struct{}{}
	seenNames := map[string]struct{}{}
	for i, artifact := range artifacts {
		if artifact.Name == "" || len(artifact.Raw) == 0 {
			return ReviewFreezeManifest{}, ErrInvalidReviewFreeze
		}
		if _, ok := seenNames[artifact.Name]; ok {
			return ReviewFreezeManifest{}, fmt.Errorf("%w: duplicate evidence name %q", ErrInvalidReviewFreeze, artifact.Name)
		}
		seenNames[artifact.Name] = struct{}{}
		if err := artifact.Window.Validate(); err != nil {
			return ReviewFreezeManifest{}, fmt.Errorf("%w: %s: %v", ErrInvalidReviewFreeze, artifact.Name, err)
		}
		if artifact.Window.RepositorySHA != repositorySHA {
			return ReviewFreezeManifest{}, fmt.Errorf("%w: evidence repository SHA mismatch", ErrInvalidReviewFreeze)
		}
		if i == 0 {
			manifest.NetworkID = artifact.Window.NetworkID
			manifest.GenesisHash = artifact.Window.GenesisHash
			manifest.CarrotPolicyHash = artifact.Window.CarrotPolicyHash
		} else if artifact.Window.NetworkID != manifest.NetworkID || artifact.Window.GenesisHash != manifest.GenesisHash || artifact.Window.CarrotPolicyHash != manifest.CarrotPolicyHash {
			return ReviewFreezeManifest{}, fmt.Errorf("%w: evidence commitments disagree", ErrInvalidReviewFreeze)
		}
		if artifact.Window.NetworkMapHash != "" {
			maps[artifact.Window.NetworkMapHash] = struct{}{}
		}
		sum := sha256.Sum256(artifact.Raw)
		manifest.Evidence = append(manifest.Evidence, EvidenceArtifactDigest{Name: artifact.Name, SHA256: hex.EncodeToString(sum[:]), EvidenceHash: artifact.Window.Hash})
	}
	for value := range maps {
		manifest.NetworkMapHashes = append(manifest.NetworkMapHashes, value)
	}
	sort.Strings(manifest.NetworkMapHashes)
	sort.Slice(manifest.Evidence, func(i, j int) bool { return manifest.Evidence[i].Name < manifest.Evidence[j].Name })
	manifest.Hash = reviewFreezeHash(manifest)
	return manifest, nil
}

func (m ReviewFreezeManifest) Validate() error {
	if m.Version != ReviewFreezeVersion || m.RepositorySHA == "" || len(m.ImageDigests) == 0 || m.NetworkID == "" || m.GenesisHash == "" || m.CarrotPolicyHash == "" || m.CreatedAt.IsZero() || len(m.Evidence) == 0 || m.Hash == "" || m.Hash != reviewFreezeHash(m) {
		return ErrInvalidReviewFreeze
	}
	seenDigests := map[string]struct{}{}
	for _, digest := range m.ImageDigests {
		if !validSHA256Digest(digest) {
			return ErrInvalidReviewFreeze
		}
		if _, ok := seenDigests[digest]; ok {
			return ErrInvalidReviewFreeze
		}
		seenDigests[digest] = struct{}{}
	}
	seen := map[string]struct{}{}
	for _, artifact := range m.Evidence {
		if artifact.Name == "" || len(artifact.SHA256) != 64 || len(artifact.EvidenceHash) != 64 {
			return ErrInvalidReviewFreeze
		}
		if _, ok := seen[artifact.Name]; ok {
			return ErrInvalidReviewFreeze
		}
		seen[artifact.Name] = struct{}{}
	}
	return nil
}

func validSHA256Digest(value string) bool {
	if !strings.HasPrefix(value, "sha256:") || len(value) != len("sha256:")+64 {
		return false
	}
	_, err := hex.DecodeString(strings.TrimPrefix(value, "sha256:"))
	return err == nil
}

func reviewFreezeHash(manifest ReviewFreezeManifest) string {
	manifest.Hash = ""
	raw, _ := json.Marshal(manifest)
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}
