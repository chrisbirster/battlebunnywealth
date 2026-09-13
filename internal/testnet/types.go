package testnet

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/chrisbirster/battlebunnywealth/internal/carrot"
)

const ProtocolVersion = 3

var (
	ErrWrongNetwork        = errors.New("wrong network")
	ErrInvalidHeight       = errors.New("invalid height")
	ErrInvalidPreviousHash = errors.New("invalid previous hash")
	ErrInvalidStateRoot    = errors.New("invalid state root")
	ErrInvalidCommittee    = errors.New("invalid committee")
	ErrInvalidProposer     = errors.New("invalid proposer")
	ErrInvalidSignature    = errors.New("invalid signature")
	ErrUnknownValidator    = errors.New("unknown validator")
	ErrDuplicateVote       = errors.New("duplicate vote")
	ErrEquivocation        = errors.New("validator equivocation")
	ErrInsufficientQuorum  = errors.New("insufficient quorum")
	ErrUnknownPeer         = errors.New("unknown peer")
	ErrReplay              = errors.New("replayed peer message")
	ErrExpiredMessage      = errors.New("expired peer message")
)

type ProtocolConfig struct {
	CommitteeTarget            int   `json:"committeeTarget"`
	BootstrapMaxValidators     int   `json:"bootstrapMaxValidators"`
	BootstrapQuorumNumerator   int   `json:"bootstrapQuorumNumerator"`
	BootstrapQuorumDenominator int   `json:"bootstrapQuorumDenominator"`
	QuorumNumerator            int   `json:"quorumNumerator"`
	QuorumDenominator          int   `json:"quorumDenominator"`
	MaxRequestBodyBytes        int64 `json:"maxRequestBodyBytes"`
	PeerReplayWindowSeconds    int64 `json:"peerReplayWindowSeconds"`
	MaxConcurrentRequestsPerIP int   `json:"maxConcurrentRequestsPerIp"`
}

func DefaultProtocolConfig() ProtocolConfig {
	return ProtocolConfig{CommitteeTarget: 64, BootstrapMaxValidators: 15, BootstrapQuorumNumerator: 3, BootstrapQuorumDenominator: 4, QuorumNumerator: 2, QuorumDenominator: 3, MaxRequestBodyBytes: 1 << 20, PeerReplayWindowSeconds: 300, MaxConcurrentRequestsPerIP: 8}
}

type Validator struct {
	ID            string    `json:"id"`
	AccountID     string    `json:"accountId"`
	DeviceID      string    `json:"deviceId"`
	Algorithm     string    `json:"algorithm"`
	PublicKey     string    `json:"publicKey"`
	RewardAddress string    `json:"rewardAddress,omitempty"`
	Authority     int64     `json:"authority"`
	Genesis       bool      `json:"genesis"`
	Active        bool      `json:"active"`
	ActivatedAt   time.Time `json:"activatedAt"`
}

type Genesis struct {
	Version          int            `json:"version"`
	NetworkID        string         `json:"networkId"`
	CreatedAt        time.Time      `json:"createdAt"`
	Config           ProtocolConfig `json:"config"`
	Validators       []Validator    `json:"validators"`
	CarrotPolicyHash string         `json:"carrotPolicyHash"`
	Hash             string         `json:"hash"`
}

func NewGenesis(networkID string, createdAt time.Time, cfg ProtocolConfig, validators []Validator) (Genesis, error) {
	if networkID == "" { return Genesis{}, errors.New("network id required") }
	if len(validators) == 0 { return Genesis{}, errors.New("at least one genesis validator required") }
	normalized := append([]Validator(nil), validators...)
	sort.Slice(normalized, func(i, j int) bool { return normalized[i].ID < normalized[j].ID })
	seen := map[string]struct{}{}
	for i := range normalized {
		v := &normalized[i]
		if v.ID == "" || v.PublicKey == "" { return Genesis{}, errors.New("validator id and public key required") }
		if _, ok := seen[v.ID]; ok { return Genesis{}, fmt.Errorf("duplicate validator %q", v.ID) }
		seen[v.ID] = struct{}{}
		v.Genesis = true; v.Active = true
		if v.ActivatedAt.IsZero() { v.ActivatedAt = createdAt.UTC() }
		if v.Authority <= 0 { v.Authority = 1 }
	}
	policy := carrot.DefaultPolicy()
	if err := policy.Validate(); err != nil { return Genesis{}, fmt.Errorf("invalid CARROT policy: %w", err) }
	g := Genesis{Version: ProtocolVersion, NetworkID: networkID, CreatedAt: createdAt.UTC(), Config: cfg, Validators: normalized, CarrotPolicyHash: policy.Hash()}
	g.Hash = hashGenesis(g)
	return g, nil
}

func (g Genesis) Validate() error {
	if g.Version != ProtocolVersion { return fmt.Errorf("unsupported protocol version %d", g.Version) }
	if g.NetworkID == "" || g.Hash == "" { return errors.New("incomplete genesis") }
	policy := carrot.DefaultPolicy()
	if err := policy.Validate(); err != nil { return fmt.Errorf("invalid local CARROT policy: %w", err) }
	if g.CarrotPolicyHash == "" || g.CarrotPolicyHash != policy.Hash() { return errors.New("CARROT policy hash mismatch") }
	if hashGenesis(g) != g.Hash { return errors.New("genesis hash mismatch") }
	if len(g.Validators) == 0 { return errors.New("empty validator set") }
	return nil
}

func hashGenesis(g Genesis) string { g.Hash = ""; raw, _ := json.Marshal(g); sum := sha256.Sum256(raw); return hex.EncodeToString(sum[:]) }

type Operation struct { Type string `json:"type"`; Key string `json:"key,omitempty"`; Value string `json:"value,omitempty"` }
type Block struct { Version int `json:"version"`; NetworkID string `json:"networkId"`; Height uint64 `json:"height"`; Round uint32 `json:"round"`; PreviousHash string `json:"previousHash"`; PreviousState string `json:"previousState"`; StateRoot string `json:"stateRoot"`; CommitteeHash string `json:"committeeHash"`; ProposerID string `json:"proposerId"`; TimestampUnix int64 `json:"timestampUnix"`; Operations []Operation `json:"operations"`; Hash string `json:"hash"` }
type Proposal struct { Block Block `json:"block"`; Signature string `json:"signature"` }
type Vote struct { NetworkID string `json:"networkId"`; Height uint64 `json:"height"`; Round uint32 `json:"round"`; BlockHash string `json:"blockHash"`; ValidatorID string `json:"validatorId"`; Decision string `json:"decision"`; Signature string `json:"signature"` }
type FinalityCertificate struct { Height uint64 `json:"height"`; Round uint32 `json:"round"`; BlockHash string `json:"blockHash"`; CommitteeHash string `json:"committeeHash"`; Quorum int `json:"quorum"`; Votes []Vote `json:"votes"` }
type FinalizedBlock struct { Block Block `json:"block"`; ProposalSignature string `json:"proposalSignature"`; Certificate FinalityCertificate `json:"certificate"` }
type Status struct { NetworkID string `json:"networkId"`; GenesisHash string `json:"genesisHash"`; Height uint64 `json:"height"`; FinalizedHash string `json:"finalizedHash"`; StateRoot string `json:"stateRoot"`; CommitteeSize int `json:"committeeSize"`; Quorum int `json:"quorum"`; PendingProposal string `json:"pendingProposal,omitempty"` }
