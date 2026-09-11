package proofofplay

import "time"

type AttestationProvider string

const (
	AttestationApple       AttestationProvider = "apple-app-attest"
	AttestationGoogle      AttestationProvider = "google-play-integrity"
	AttestationDevelopment AttestationProvider = "development"
)

type Participant struct {
	ID                  string              `json:"id"`
	DID                 string              `json:"did"`
	DevicePublicKey     string              `json:"devicePublicKey"`
	AttestationProvider AttestationProvider `json:"attestationProvider"`
	Reputation          uint64              `json:"reputation"`
	JoinedAt            time.Time           `json:"joinedAt"`
}

type Challenge struct {
	ID        string    `json:"id"`
	Epoch     uint64    `json:"epoch"`
	Nonce     string    `json:"nonce"`
	IssuedAt  time.Time `json:"issuedAt"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type ParticipationProof struct {
	Version             uint16              `json:"version"`
	ParticipantID       string              `json:"participantId"`
	Epoch               uint64              `json:"epoch"`
	ChallengeID         string              `json:"challengeId"`
	DevicePublicKey     string              `json:"devicePublicKey"`
	AttestationProvider AttestationProvider `json:"attestationProvider"`
	AttestationDigest   string              `json:"attestationDigest"`
	Signature           string              `json:"signature"`
	SubmittedAt         time.Time           `json:"submittedAt"`
}

type Transaction struct {
	ID      string            `json:"id"`
	Type    string            `json:"type"`
	Payload map[string]string `json:"payload,omitempty"`
}
