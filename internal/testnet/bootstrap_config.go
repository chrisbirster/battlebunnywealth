package testnet

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"
)

var bootstrapNodeNameRE = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)

// BootstrapNode is the public identity and endpoint of one independently
// deployed full node. It is infrastructure identity only and grants no
// consensus voting power.
type BootstrapNode struct {
	Name      string `json:"name"`
	NodeID    string `json:"nodeId"`
	PublicKey string `json:"publicKey"`
	URL       string `json:"url"`
}

// BuildNodeConfigs deterministically builds one NodeConfig per full node.
// Every config receives the exact same genesis and all other nodes as peers.
func BuildNodeConfigs(genesis Genesis, nodes []BootstrapNode, syncIntervalSeconds int) (map[string]NodeConfig, error) {
	if err := genesis.Validate(); err != nil {
		return nil, fmt.Errorf("invalid genesis: %w", err)
	}
	if len(nodes) < 2 {
		return nil, errors.New("at least two bootstrap nodes required")
	}
	if syncIntervalSeconds <= 0 {
		syncIntervalSeconds = 5
	}

	normalized := append([]BootstrapNode(nil), nodes...)
	sort.Slice(normalized, func(i, j int) bool { return normalized[i].Name < normalized[j].Name })

	seenNames := map[string]struct{}{}
	seenIDs := map[string]struct{}{}
	seenURLs := map[string]struct{}{}
	for i := range normalized {
		n := &normalized[i]
		n.Name = strings.TrimSpace(strings.ToLower(n.Name))
		n.NodeID = strings.TrimSpace(strings.ToLower(n.NodeID))
		n.PublicKey = strings.TrimSpace(n.PublicKey)
		n.URL = strings.TrimRight(strings.TrimSpace(n.URL), "/")
		if !bootstrapNodeNameRE.MatchString(n.Name) {
			return nil, fmt.Errorf("invalid node name %q", n.Name)
		}
		if _, ok := seenNames[n.Name]; ok {
			return nil, fmt.Errorf("duplicate node name %q", n.Name)
		}
		seenNames[n.Name] = struct{}{}
		if _, ok := seenIDs[n.NodeID]; ok {
			return nil, fmt.Errorf("duplicate node id %q", n.NodeID)
		}
		seenIDs[n.NodeID] = struct{}{}
		if _, ok := seenURLs[n.URL]; ok {
			return nil, fmt.Errorf("duplicate node url %q", n.URL)
		}
		seenURLs[n.URL] = struct{}{}

		raw, err := base64.RawURLEncoding.DecodeString(n.PublicKey)
		if err != nil || len(raw) != ed25519.PublicKeySize {
			return nil, fmt.Errorf("invalid public key for %q", n.Name)
		}
		sum := sha256.Sum256(raw)
		wantID := hex.EncodeToString(sum[:16])
		if n.NodeID != wantID {
			return nil, fmt.Errorf("node id/public key mismatch for %q: got %s want %s", n.Name, n.NodeID, wantID)
		}
		parsed, err := url.Parse(n.URL)
		if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.Path != "" {
			return nil, fmt.Errorf("node %q requires a bare https URL", n.Name)
		}
	}

	configs := make(map[string]NodeConfig, len(normalized))
	for _, current := range normalized {
		peers := make([]Peer, 0, len(normalized)-1)
		for _, other := range normalized {
			if other.NodeID == current.NodeID {
				continue
			}
			peers = append(peers, Peer{NodeID: other.NodeID, PublicKey: other.PublicKey, URL: other.URL})
		}
		configs[current.Name] = NodeConfig{Listen: ":9101", Genesis: genesis, Peers: peers, SyncIntervalSeconds: syncIntervalSeconds}
	}
	return configs, nil
}
