package publictestnet

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"sort"
	"strings"
)

const NetworkMapVersion = 1

var ErrInvalidNetworkMap = errors.New("invalid public-testnet network metadata map")

type NetworkMapEntry struct {
	CIDR     string `json:"cidr"`
	ASN      uint32 `json:"asn,omitempty"`
	Provider string `json:"provider,omitempty"`
}

type NetworkMapDocument struct {
	Version int               `json:"version"`
	Entries []NetworkMapEntry `json:"entries"`
	Hash    string            `json:"hash"`
}

type compiledNetworkEntry struct {
	network  *net.IPNet
	prefix   int
	metadata PeerNetworkMetadata
}

type NetworkMapClassifier struct {
	document NetworkMapDocument
	entries  []compiledNetworkEntry
}

func LoadNetworkMap(path string) (*NetworkMapClassifier, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseNetworkMap(raw)
}

func ParseNetworkMap(raw []byte) (*NetworkMapClassifier, error) {
	var document NetworkMapDocument
	if err := json.Unmarshal(raw, &document); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidNetworkMap, err)
	}
	if document.Version != NetworkMapVersion || len(document.Entries) == 0 {
		return nil, ErrInvalidNetworkMap
	}
	normalized := make([]NetworkMapEntry, 0, len(document.Entries))
	compiled := make([]compiledNetworkEntry, 0, len(document.Entries))
	seen := map[string]struct{}{}
	for _, entry := range document.Entries {
		_, network, err := net.ParseCIDR(strings.TrimSpace(entry.CIDR))
		if err != nil || network == nil || (entry.ASN == 0 && strings.TrimSpace(entry.Provider) == "") {
			return nil, ErrInvalidNetworkMap
		}
		canonical := network.String()
		if _, ok := seen[canonical]; ok {
			return nil, ErrInvalidNetworkMap
		}
		seen[canonical] = struct{}{}
		provider := strings.TrimSpace(entry.Provider)
		ones, _ := network.Mask.Size()
		normalized = append(normalized, NetworkMapEntry{CIDR: canonical, ASN: entry.ASN, Provider: provider})
		compiled = append(compiled, compiledNetworkEntry{network: network, prefix: ones, metadata: PeerNetworkMetadata{ASN: entry.ASN, Provider: provider}})
	}
	sort.Slice(normalized, func(i, j int) bool {
		if normalized[i].CIDR == normalized[j].CIDR {
			if normalized[i].ASN == normalized[j].ASN {
				return normalized[i].Provider < normalized[j].Provider
			}
			return normalized[i].ASN < normalized[j].ASN
		}
		return normalized[i].CIDR < normalized[j].CIDR
	})
	document.Entries = normalized
	calculated := networkMapHash(document)
	if document.Hash != "" && document.Hash != calculated {
		return nil, fmt.Errorf("%w: hash mismatch", ErrInvalidNetworkMap)
	}
	document.Hash = calculated
	sort.Slice(compiled, func(i, j int) bool { return compiled[i].prefix > compiled[j].prefix })
	return &NetworkMapClassifier{document: document, entries: compiled}, nil
}

func NewNetworkMapDocument(entries []NetworkMapEntry) (NetworkMapDocument, error) {
	raw, err := json.Marshal(NetworkMapDocument{Version: NetworkMapVersion, Entries: entries})
	if err != nil {
		return NetworkMapDocument{}, err
	}
	classifier, err := ParseNetworkMap(raw)
	if err != nil {
		return NetworkMapDocument{}, err
	}
	return classifier.Document(), nil
}

func (c *NetworkMapClassifier) Classify(ip net.IP) PeerNetworkMetadata {
	if c == nil || ip == nil {
		return PeerNetworkMetadata{}
	}
	for _, entry := range c.entries {
		if entry.network.Contains(ip) {
			return entry.metadata
		}
	}
	return PeerNetworkMetadata{}
}

func (c *NetworkMapClassifier) Document() NetworkMapDocument {
	if c == nil {
		return NetworkMapDocument{}
	}
	copy := c.document
	copy.Entries = append([]NetworkMapEntry(nil), c.document.Entries...)
	return copy
}

func (c *NetworkMapClassifier) Hash() string {
	if c == nil {
		return ""
	}
	return c.document.Hash
}

func networkMapHash(document NetworkMapDocument) string {
	document.Hash = ""
	raw, _ := json.Marshal(document)
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}
