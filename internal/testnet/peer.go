package testnet

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"
)

type Peer struct {
	NodeID    string `json:"nodeId"`
	URL       string `json:"url"`
	PublicKey string `json:"publicKey"`
}
type PeerEnvelope struct {
	NetworkID     string          `json:"networkId"`
	NodeID        string          `json:"nodeId"`
	TimestampUnix int64           `json:"timestampUnix"`
	Nonce         string          `json:"nonce"`
	Kind          string          `json:"kind"`
	Payload       json.RawMessage `json:"payload"`
	Signature     string          `json:"signature"`
}

func SignPeerEnvelope(networkID, kind, nonce string, key NodeKey, payload any, now time.Time) (PeerEnvelope, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return PeerEnvelope{}, err
	}
	e := PeerEnvelope{NetworkID: networkID, NodeID: key.ID(), TimestampUnix: now.UTC().Unix(), Nonce: nonce, Kind: kind, Payload: raw}
	e.Signature = base64.RawURLEncoding.EncodeToString(ed25519.Sign(key.Private, []byte(peerSigningMessage(e))))
	return e, nil
}
func peerSigningMessage(e PeerEnvelope) string {
	sum := sha256.Sum256(e.Payload)
	return fmt.Sprintf("bbw-pop-peer/v1\nnetwork=%s\nnode=%s\ntime=%d\nnonce=%s\nkind=%s\npayload=%s", e.NetworkID, e.NodeID, e.TimestampUnix, e.Nonce, e.Kind, hex.EncodeToString(sum[:]))
}

type PeerGuard struct {
	mu           sync.Mutex
	networkID    string
	replayWindow time.Duration
	peers        map[string]ed25519.PublicKey
	seen         map[string]time.Time
	rejected     map[string]int64
	accepted     map[string]int64
}

func NewPeerGuard(networkID string, replayWindow time.Duration, peers []Peer) (*PeerGuard, error) {
	g := &PeerGuard{networkID: networkID, replayWindow: replayWindow, peers: map[string]ed25519.PublicKey{}, seen: map[string]time.Time{}, rejected: map[string]int64{}, accepted: map[string]int64{}}
	if g.replayWindow <= 0 {
		g.replayWindow = 5 * time.Minute
	}
	for _, peer := range peers {
		raw, err := base64.RawURLEncoding.DecodeString(peer.PublicKey)
		if err != nil || len(raw) != ed25519.PublicKeySize {
			return nil, fmt.Errorf("invalid peer %q public key", peer.NodeID)
		}
		g.peers[peer.NodeID] = ed25519.PublicKey(raw)
	}
	return g, nil
}
func (g *PeerGuard) Verify(e PeerEnvelope, now time.Time) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	reject := func(err error) error { g.rejected[e.NodeID]++; return err }
	if e.NetworkID != g.networkID {
		return reject(ErrWrongNetwork)
	}
	pub, ok := g.peers[e.NodeID]
	if !ok {
		return reject(ErrUnknownPeer)
	}
	ts := time.Unix(e.TimestampUnix, 0)
	if now.Sub(ts) > g.replayWindow || ts.Sub(now) > g.replayWindow {
		return reject(ErrExpiredMessage)
	}
	if e.Nonce == "" {
		return reject(errors.New("missing peer nonce"))
	}
	key := e.NodeID + "/" + e.Nonce
	if _, ok := g.seen[key]; ok {
		return reject(ErrReplay)
	}
	sig, err := base64.RawURLEncoding.DecodeString(e.Signature)
	if err != nil || !ed25519.Verify(pub, []byte(peerSigningMessage(e)), sig) {
		return reject(ErrInvalidSignature)
	}
	g.seen[key] = now
	for k, t := range g.seen {
		if now.Sub(t) > g.replayWindow {
			delete(g.seen, k)
		}
	}
	g.accepted[e.NodeID]++
	return nil
}
func (g *PeerGuard) Metrics() map[string]map[string]int64 {
	g.mu.Lock()
	defer g.mu.Unlock()
	out := map[string]map[string]int64{}
	ids := make([]string, 0, len(g.peers))
	for id := range g.peers {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		out[id] = map[string]int64{"accepted": g.accepted[id], "rejected": g.rejected[id]}
	}
	return out
}
