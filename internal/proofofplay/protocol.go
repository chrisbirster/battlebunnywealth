package proofofplay

import (
	"errors"
	"fmt"
	"time"
)

type Config struct {
	EpochDuration             time.Duration `json:"-"`
	ChallengeTTL              time.Duration `json:"-"`
	CommitteeSize             int           `json:"committeeSize"`
	QuorumNumerator           int           `json:"quorumNumerator"`
	QuorumDenominator         int           `json:"quorumDenominator"`
	MaxHumanChallengesPerDay  int           `json:"maxHumanChallengesPerDay"`
	SecondDeviceWeightPercent int           `json:"secondDeviceWeightPercent"`
	ThirdDeviceWeightPercent  int           `json:"thirdDeviceWeightPercent"`
	LaterDeviceWeightPercent  int           `json:"laterDeviceWeightPercent"`
}

func DefaultConfig() Config {
	return Config{
		EpochDuration:             10 * time.Minute,
		ChallengeTTL:              5 * time.Minute,
		CommitteeSize:             64,
		QuorumNumerator:           2,
		QuorumDenominator:         3,
		MaxHumanChallengesPerDay:  4,
		SecondDeviceWeightPercent: 25,
		ThirdDeviceWeightPercent:  10,
		LaterDeviceWeightPercent:  2,
	}
}

type AttestationVerifier interface { Verify(participant Participant, challenge Challenge, proof ParticipationProof) error }
type SignatureVerifier interface { Verify(participant Participant, challenge Challenge, proof ParticipationProof) error }

type Protocol struct { config Config; chain *Chain; started time.Time }
func NewProtocol(config Config, chain *Chain) *Protocol { return &Protocol{config: config, chain: chain, started: time.Now().UTC()} }
func (p *Protocol) CurrentEpoch(now time.Time) uint64 { if now.Before(p.started) { return 0 }; return uint64(now.Sub(p.started) / p.config.EpochDuration) }

type NetworkStatus struct {
	Protocol                   string   `json:"protocol"`
	Phase                      string   `json:"phase"`
	Epoch                      uint64   `json:"epoch"`
	ChainHeight                uint64   `json:"chainHeight"`
	CommitteeSize              int      `json:"committeeSize"`
	Quorum                     string   `json:"quorum"`
	EpochSeconds               int64    `json:"epochSeconds"`
	ChallengeTTLSeconds        int64    `json:"challengeTtlSeconds"`
	MaxHumanChallengesPerDay   int      `json:"maxHumanChallengesPerDay"`
	AttestationProviders       []string `json:"attestationProviders"`
	AuthorityMode              string   `json:"authorityMode"`
	ProductionCommitteeEnabled bool     `json:"productionCommitteeEnabled"`
}

func (p *Protocol) Status(now time.Time) NetworkStatus {
	return NetworkStatus{
		Protocol: "proof-of-play/0.7",
		Phase: "attested-participation-prototype",
		Epoch: p.CurrentEpoch(now), ChainHeight: p.chain.Height(), CommitteeSize: p.config.CommitteeSize,
		Quorum: fmt.Sprintf("%d/%d", p.config.QuorumNumerator, p.config.QuorumDenominator),
		EpochSeconds: int64(p.config.EpochDuration / time.Second), ChallengeTTLSeconds: int64(p.config.ChallengeTTL / time.Second),
		MaxHumanChallengesPerDay: p.config.MaxHumanChallengesPerDay,
		AttestationProviders: []string{ProviderAppleAppAttest, ProviderGooglePlayIntegrity},
		AuthorityMode: "bounded-decaying-attestation-gated-prototype",
		// Distributed production consensus still does not exist; v0.7 only defines a permissioned-testnet eligibility gate.
		ProductionCommitteeEnabled: false,
	}
}
func (p *Protocol) Head() Block { return p.chain.Head() }

var (
	ErrChallengeExpired    = errors.New("challenge expired")
	ErrEpochMismatch       = errors.New("proof epoch does not match challenge")
	ErrChallengeMismatch   = errors.New("proof challenge id does not match challenge")
	ErrParticipantMismatch = errors.New("proof participant does not match participant")
	ErrDeviceMismatch      = errors.New("proof device key does not match enrolled device")
)

func ValidateEnvelope(now time.Time, participant Participant, challenge Challenge, proof ParticipationProof) error {
	if now.After(challenge.ExpiresAt) { return ErrChallengeExpired }
	if proof.Epoch != challenge.Epoch { return ErrEpochMismatch }
	if proof.ChallengeID != challenge.ID { return ErrChallengeMismatch }
	if proof.ParticipantID != participant.ID { return ErrParticipantMismatch }
	if proof.DevicePublicKey != participant.DevicePublicKey { return ErrDeviceMismatch }
	return nil
}
