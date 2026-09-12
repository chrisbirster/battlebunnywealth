package testnet

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Relayer struct {
	NetworkID string
	Key       NodeKey
	Peers     []Peer
	Client    *http.Client
}

func (r *Relayer) Broadcast(ctx context.Context, kind string, payload any) map[string]error {
	results := map[string]error{}
	if r == nil {
		return results
	}
	client := r.Client
	if client == nil {
		client = &http.Client{Timeout: 3 * time.Second}
	}
	suffix := ""
	switch kind {
	case "proposal":
		suffix = "/v1/peer/proposals"
	case "vote":
		suffix = "/v1/peer/votes"
	default:
		results["local"] = fmt.Errorf("unsupported relay kind %q", kind)
		return results
	}
	for _, peer := range r.Peers {
		nonceBytes := make([]byte, 18)
		if _, err := rand.Read(nonceBytes); err != nil {
			results[peer.NodeID] = err
			continue
		}
		envelope, err := SignPeerEnvelope(r.NetworkID, kind, base64.RawURLEncoding.EncodeToString(nonceBytes), r.Key, payload, time.Now().UTC())
		if err != nil {
			results[peer.NodeID] = err
			continue
		}
		raw, err := json.Marshal(envelope)
		if err != nil {
			results[peer.NodeID] = err
			continue
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(peer.URL, "/")+suffix, bytes.NewReader(raw))
		if err != nil {
			results[peer.NodeID] = err
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		if err != nil {
			results[peer.NodeID] = err
			continue
		}
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64<<10))
		_ = resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			results[peer.NodeID] = fmt.Errorf("relay status %d", resp.StatusCode)
		}
	}
	return results
}
