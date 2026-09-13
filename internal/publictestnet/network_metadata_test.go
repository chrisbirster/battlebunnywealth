package publictestnet

import (
	"encoding/json"
	"errors"
	"net"
	"testing"
)

func TestNetworkMapClassifierUsesLongestPrefixAndStableHash(t *testing.T) {
	document, err := NewNetworkMapDocument([]NetworkMapEntry{
		{CIDR: "203.0.113.0/24", ASN: 64500, Provider: "cloud-a"},
		{CIDR: "203.0.113.128/25", ASN: 64501, Provider: "cloud-b"},
		{CIDR: "2001:db8::/32", ASN: 64510, Provider: "v6-provider"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if document.Hash == "" {
		t.Fatal("network map hash missing")
	}
	raw, _ := json.Marshal(document)
	classifier, err := ParseNetworkMap(raw)
	if err != nil {
		t.Fatal(err)
	}
	if classifier.Hash() != document.Hash {
		t.Fatalf("hash=%s want=%s", classifier.Hash(), document.Hash)
	}
	if got := classifier.Classify(net.ParseIP("203.0.113.200")); got.ASN != 64501 || got.Provider != "cloud-b" {
		t.Fatalf("longest-prefix metadata=%+v", got)
	}
	if got := classifier.Classify(net.ParseIP("203.0.113.20")); got.ASN != 64500 || got.Provider != "cloud-a" {
		t.Fatalf("/24 metadata=%+v", got)
	}
	if got := classifier.Classify(net.ParseIP("2001:db8::1")); got.ASN != 64510 {
		t.Fatalf("ipv6 metadata=%+v", got)
	}
	if got := classifier.Classify(net.ParseIP("198.51.100.1")); got.ASN != 0 || got.Provider != "" {
		t.Fatalf("unknown metadata=%+v", got)
	}
}

func TestNetworkMapRejectsTamperedHash(t *testing.T) {
	document, err := NewNetworkMapDocument([]NetworkMapEntry{{CIDR: "203.0.113.0/24", ASN: 64500, Provider: "cloud-a"}})
	if err != nil {
		t.Fatal(err)
	}
	document.Entries[0].Provider = "tampered"
	raw, _ := json.Marshal(document)
	if _, err := ParseNetworkMap(raw); !errors.Is(err, ErrInvalidNetworkMap) {
		t.Fatalf("tampered map err=%v", err)
	}
}
