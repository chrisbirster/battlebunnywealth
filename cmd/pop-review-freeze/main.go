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

type evidenceFiles []string
func (e *evidenceFiles) String() string { return strings.Join(*e,",") }
func (e *evidenceFiles) Set(value string) error { value=strings.TrimSpace(value); if value=="" { return fmt.Errorf("empty evidence path") }; *e=append(*e,value); return nil }

type evidenceEnvelope struct { Evidence publictestnet.LiveEvidenceWindow `json:"evidence"` }

func main() {
	repositorySHA := flag.String("repo-sha","","exact merged dev commit SHA to freeze")
	outPath := flag.String("out","review-freeze.json","output manifest path")
	var files evidenceFiles; flag.Var(&files,"evidence","live evidence JSON file; repeat for every retained evidence window")
	flag.Parse()
	if strings.TrimSpace(*repositorySHA)=="" || len(files)==0 { fatal(fmt.Errorf("-repo-sha and at least one -evidence are required")) }
	artifacts := make([]publictestnet.NamedEvidenceWindow,0,len(files))
	for _, path := range files {
		raw,err:=os.ReadFile(path); if err!=nil { fatal(err) }
		var envelope evidenceEnvelope
		if err:=json.Unmarshal(raw,&envelope); err!=nil { fatal(fmt.Errorf("decode %s: %w",path,err)) }
		if envelope.Evidence.Hash=="" {
			var direct publictestnet.LiveEvidenceWindow
			if err:=json.Unmarshal(raw,&direct); err!=nil { fatal(fmt.Errorf("decode direct evidence %s: %w",path,err)) }
			envelope.Evidence=direct
		}
		artifacts=append(artifacts,publictestnet.NamedEvidenceWindow{Name:filepath.Base(path),Raw:raw,Window:envelope.Evidence})
	}
	manifest,err:=publictestnet.BuildReviewFreeze(strings.TrimSpace(*repositorySHA),artifacts,time.Now().UTC()); if err!=nil { fatal(err) }
	raw,err:=json.MarshalIndent(manifest,"","  "); if err!=nil { fatal(err) }
	if err:=os.WriteFile(*outPath,append(raw,'\n'),0o600); err!=nil { fatal(err) }
	fmt.Printf("review freeze: %s evidence=%d hash=%s\n",*outPath,len(manifest.Evidence),manifest.Hash)
}

func fatal(err error) { fmt.Fprintln(os.Stderr,"pop-review-freeze:",err); os.Exit(1) }
