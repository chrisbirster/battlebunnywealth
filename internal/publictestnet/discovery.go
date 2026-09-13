package publictestnet

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/url"
	"sort"
	"sync"
	"time"

	"github.com/chrisbirster/battlebunnywealth/internal/testnet"
)

var (
	ErrInvalidAnnouncement = errors.New("invalid public-node announcement")
	ErrPeerDirectoryFull   = errors.New("public peer directory full")
	ErrPeerHostLimit       = errors.New("public peer host limit reached")
	ErrPeerPrefixLimit     = errors.New("public peer network-prefix limit reached")
	ErrPeerASNLimit        = errors.New("public peer ASN diversity limit reached")
	ErrPeerProviderLimit   = errors.New("public peer provider diversity limit reached")
)

type NodeAnnouncement struct {
	NetworkID     string `json:"networkId"`
	NodeID        string `json:"nodeId"`
	PublicKey     string `json:"publicKey"`
	URL           string `json:"url"`
	TimestampUnix int64  `json:"timestampUnix"`
	ExpiresUnix   int64  `json:"expiresUnix"`
	Nonce         string `json:"nonce"`
	Signature     string `json:"signature"`
}

type PeerNetworkMetadata struct {
	ASN      uint32 `json:"asn,omitempty"`
	Provider string `json:"provider,omitempty"`
}

// NetworkClassifier is deliberately server-side. ASN/provider labels in a
// self-signed node announcement would be attacker-controlled and therefore
// useless for Sybil/eclipsing resistance.
type NetworkClassifier interface {
	Classify(ip net.IP) PeerNetworkMetadata
}

type NetworkClassifierFunc func(ip net.IP) PeerNetworkMetadata

func (f NetworkClassifierFunc) Classify(ip net.IP) PeerNetworkMetadata { return f(ip) }

func BuildAnnouncement(networkID, endpoint, nonce string, key testnet.NodeKey, now time.Time, ttl time.Duration) (NodeAnnouncement, error) {
	if ttl <= 0 || ttl > 24*time.Hour {
		return NodeAnnouncement{}, ErrInvalidAnnouncement
	}
	a := NodeAnnouncement{
		NetworkID:     networkID,
		NodeID:        key.ID(),
		PublicKey:     key.PublicText(),
		URL:           endpoint,
		TimestampUnix: now.UTC().Unix(),
		ExpiresUnix:   now.Add(ttl).UTC().Unix(),
		Nonce:         nonce,
	}
	if err := validateEndpoint(endpoint); err != nil {
		return NodeAnnouncement{}, err
	}
	a.Signature = base64.RawURLEncoding.EncodeToString(ed25519.Sign(key.Private, []byte(announcementMessage(a))))
	return a, nil
}

func VerifyAnnouncement(a NodeAnnouncement, networkID string, now time.Time) error {
	if a.NetworkID != networkID || a.NodeID == "" || a.Nonce == "" || a.ExpiresUnix <= a.TimestampUnix {
		return ErrInvalidAnnouncement
	}
	if err := validateEndpoint(a.URL); err != nil {
		return err
	}
	if now.Unix() < a.TimestampUnix-300 || now.Unix() > a.TimestampUnix+300 || a.ExpiresUnix <= now.Unix() || a.ExpiresUnix > now.Add(24*time.Hour).Unix() {
		return ErrInvalidAnnouncement
	}
	pub, err := base64.RawURLEncoding.DecodeString(a.PublicKey)
	if err != nil || len(pub) != ed25519.PublicKeySize || nodeID(ed25519.PublicKey(pub)) != a.NodeID {
		return ErrInvalidAnnouncement
	}
	sig, err := base64.RawURLEncoding.DecodeString(a.Signature)
	if err != nil || !ed25519.Verify(ed25519.PublicKey(pub), []byte(announcementMessage(a)), sig) {
		return ErrInvalidAnnouncement
	}
	return nil
}

func announcementMessage(a NodeAnnouncement) string {
	return fmt.Sprintf("bbw-public-node/v1\nnetwork=%s\nnode=%s\npublicKey=%s\nurl=%s\ntimestamp=%d\nexpires=%d\nnonce=%s", a.NetworkID, a.NodeID, a.PublicKey, a.URL, a.TimestampUnix, a.ExpiresUnix, a.Nonce)
}

func nodeID(pub ed25519.PublicKey) string {
	sum := sha256.Sum256(pub)
	return hex.EncodeToString(sum[:16])
}

func validateEndpoint(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || u.User != nil || u.Fragment != "" {
		return ErrInvalidAnnouncement
	}
	return nil
}

type Directory struct {
	mu          sync.Mutex
	network     string
	max         int
	perHost     int
	perPrefix   int
	perASN      int
	perProvider int
	classifier  NetworkClassifier
	peers       map[string]NodeAnnouncement
	metadata    map[string]PeerNetworkMetadata
}

func NewDirectory(networkID string, maxPeers, maxPerHost int) *Directory {
	if maxPeers <= 0 {
		maxPeers = 128
	}
	if maxPerHost <= 0 {
		maxPerHost = 4
	}
	perPrefix := 16
	if perPrefix > maxPeers {
		perPrefix = maxPeers
	}
	return &Directory{
		network:     networkID,
		max:         maxPeers,
		perHost:     maxPerHost,
		perPrefix:   perPrefix,
		perASN:      32,
		perProvider: 64,
		peers:       map[string]NodeAnnouncement{},
		metadata:    map[string]PeerNetworkMetadata{},
	}
}

// SetNetworkClassifier enables research-grade ASN/provider diversity limits.
// Production deployment should supply metadata from a trusted local database or
// resolver, never from peer-declared fields. Unknown metadata is not rejected;
// IP host/prefix limits still apply.
func (d *Directory) SetNetworkClassifier(classifier NetworkClassifier, maxPerASN, maxPerProvider int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.classifier = classifier
	if maxPerASN > 0 {
		d.perASN = maxPerASN
	}
	if maxPerProvider > 0 {
		d.perProvider = maxPerProvider
	}
}

func (d *Directory) Upsert(a NodeAnnouncement, now time.Time) error {
	if err := VerifyAnnouncement(a, d.network, now); err != nil {
		return err
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	d.pruneLocked(now)

	_, existing := d.peers[a.NodeID]
	if !existing && len(d.peers) >= d.max {
		return ErrPeerDirectoryFull
	}
	u, _ := url.Parse(a.URL)
	host := u.Hostname()
	prefix := peerPrefix(host)
	metadata := PeerNetworkMetadata{}
	if d.classifier != nil {
		if ip := net.ParseIP(host); ip != nil {
			metadata = d.classifier.Classify(ip)
		}
	}

	hostCount, prefixCount, asnCount, providerCount := 0, 0, 0, 0
	for id, peer := range d.peers {
		if id == a.NodeID {
			continue
		}
		pu, _ := url.Parse(peer.URL)
		peerHost := pu.Hostname()
		if peerHost == host {
			hostCount++
		}
		if prefix != "" && peerPrefix(peerHost) == prefix {
			prefixCount++
		}
		peerMetadata := d.metadata[id]
		if metadata.ASN != 0 && peerMetadata.ASN == metadata.ASN {
			asnCount++
		}
		if metadata.Provider != "" && peerMetadata.Provider == metadata.Provider {
			providerCount++
		}
	}
	if hostCount >= d.perHost {
		return ErrPeerHostLimit
	}
	if prefix != "" && prefixCount >= d.perPrefix {
		return ErrPeerPrefixLimit
	}
	if metadata.ASN != 0 && asnCount >= d.perASN {
		return ErrPeerASNLimit
	}
	if metadata.Provider != "" && providerCount >= d.perProvider {
		return ErrPeerProviderLimit
	}

	d.peers[a.NodeID] = a
	d.metadata[a.NodeID] = metadata
	return nil
}

func peerPrefix(host string) string {
	ip := net.ParseIP(host)
	if ip == nil {
		return ""
	}
	if v4 := ip.To4(); v4 != nil {
		return fmt.Sprintf("v4:%d.%d.%d", v4[0], v4[1], v4[2])
	}
	v6 := ip.To16()
	if v6 == nil {
		return ""
	}
	return fmt.Sprintf("v6:%x%x%x%x%x%x", v6[0], v6[1], v6[2], v6[3], v6[4], v6[5])
}

func (d *Directory) List(now time.Time) []NodeAnnouncement {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.pruneLocked(now)
	out := make([]NodeAnnouncement, 0, len(d.peers))
	for _, peer := range d.peers {
		out = append(out, peer)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].NodeID < out[j].NodeID })
	return out
}

func (d *Directory) Metadata(nodeID string) (PeerNetworkMetadata, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	metadata, ok := d.metadata[nodeID]
	return metadata, ok
}

func (d *Directory) pruneLocked(now time.Time) {
	for id, peer := range d.peers {
		if peer.ExpiresUnix <= now.Unix() {
			delete(d.peers, id)
			delete(d.metadata, id)
		}
	}
}
