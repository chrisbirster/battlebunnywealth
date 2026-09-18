package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chrisbirster/battlebunnywealth/internal/publictestnet"
)

type repeatedFlag []string

func (f *repeatedFlag) String() string { return strings.Join(*f, ",") }
func (f *repeatedFlag) Set(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("empty value")
	}
	*f = append(*f, value)
	return nil
}

type evidenceEnvelope struct {
	Evidence publictestnet.LiveEvidenceWindow `json:"evidence"`
}

func main() {
	repositorySHA := flag.String("repo-sha", "", "exact merged dev commit SHA to freeze")
	flyTopologyPath := flag.String("fly-topology", "", "validated Fly topology evidence JSON for the deployed review target")
	outPath := flag.String("out", "review-freeze.json", "output manifest path")
	var files repeatedFlag
	var imageDigests repeatedFlag
	var supportingFiles repeatedFlag
	flag.Var(&files, "evidence", "live evidence JSON file; repeat for every retained evidence window")
	flag.Var(&imageDigests, "image-digest", "deployed immutable container image digest (sha256:...); repeat when operators used more than one identical-code image build")
	flag.Var(&supportingFiles, "supporting-evidence", "supporting evidence/provenance file to hash into the freeze; repeat as needed")
	flag.Parse()
	if strings.TrimSpace(*repositorySHA) == "" || strings.TrimSpace(*flyTopologyPath) == "" || len(files) == 0 || len(imageDigests) == 0 {
		fatal(fmt.Errorf("-repo-sha, -fly-topology, at least one -image-digest, and at least one -evidence are required"))
	}
	artifacts := make([]publictestnet.NamedEvidenceWindow, 0, len(files))
	for _, path := range files {
		raw, err := os.ReadFile(path)
		if err != nil {
			fatal(err)
		}
		var envelope evidenceEnvelope
		if err := json.Unmarshal(raw, &envelope); err != nil {
			fatal(fmt.Errorf("decode %s: %w", path, err))
		}
		if envelope.Evidence.Hash == "" {
			var direct publictestnet.LiveEvidenceWindow
			if err := json.Unmarshal(raw, &direct); err != nil {
				fatal(fmt.Errorf("decode direct evidence %s: %w", path, err))
			}
			envelope.Evidence = direct
		}
		artifacts = append(artifacts, publictestnet.NamedEvidenceWindow{Name: filepath.Base(path), Raw: raw, Window: envelope.Evidence})
	}
	topologyRaw, err := os.ReadFile(*flyTopologyPath)
	if err != nil {
		fatal(err)
	}
	var topology publictestnet.FlyTopologyEvidence
	if err := json.Unmarshal(topologyRaw, &topology); err != nil {
		fatal(fmt.Errorf("decode Fly topology %s: %w", *flyTopologyPath, err))
	}
	if err := topology.Validate(); err != nil {
		fatal(fmt.Errorf("validate Fly topology %s: %w", *flyTopologyPath, err))
	}
	if topology.RepositorySHA != strings.TrimSpace(*repositorySHA) {
		fatal(fmt.Errorf("Fly topology repository SHA mismatch"))
	}
	if !sameDigestSet(topology.ImageDigests, imageDigests) {
		fatal(fmt.Errorf("Fly topology image digests do not match -image-digest values"))
	}

	supporting := make([]publictestnet.NamedSupportingArtifact, 0, len(supportingFiles)+1)
	supporting = append(supporting, publictestnet.NamedSupportingArtifact{Name: filepath.Base(*flyTopologyPath), Raw: topologyRaw})
	for _, path := range supportingFiles {
		raw, err := os.ReadFile(path)
		if err != nil {
			fatal(err)
		}
		supporting = append(supporting, publictestnet.NamedSupportingArtifact{Name: filepath.Base(path), Raw: raw})
	}
	manifest, err := publictestnet.BuildReviewFreezeWithSupporting(strings.TrimSpace(*repositorySHA), artifacts, supporting, time.Now().UTC(), imageDigests...)
	if err != nil {
		fatal(err)
	}
	raw, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		fatal(err)
	}
	if err := os.WriteFile(*outPath, append(raw, '\n'), 0o600); err != nil {
		fatal(err)
	}
	fmt.Printf("review freeze: %s evidence=%d supporting=%d images=%d hash=%s\n", *outPath, len(manifest.Evidence), len(manifest.SupportingEvidence), len(manifest.ImageDigests), manifest.Hash)
}

func sameDigestSet(want []string, got []string) bool {
	wanted := map[string]struct{}{}
	for _, digest := range want {
		wanted[strings.ToLower(strings.TrimSpace(digest))] = struct{}{}
	}
	actual := map[string]struct{}{}
	for _, digest := range got {
		actual[strings.ToLower(strings.TrimSpace(digest))] = struct{}{}
	}
	if len(wanted) != len(actual) {
		return false
	}
	for digest := range wanted {
		if _, ok := actual[digest]; !ok {
			return false
		}
	}
	return true
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "pop-review-freeze:", err)
	os.Exit(1)
}
