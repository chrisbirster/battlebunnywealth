package identity

import "time"

const SchemaVersion = 1

const (
	DeviceStatusActive  = "active"
	DeviceStatusRevoked = "revoked"
)

type PasskeyCredential struct {
	ID         string    `json:"id"`
	PublicKeyX string    `json:"publicKeyX"`
	PublicKeyY string    `json:"publicKeyY"`
	SignCount  uint32    `json:"signCount"`
	Transports []string  `json:"transports,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
	LastUsedAt time.Time `json:"lastUsedAt,omitempty"`
}

type ATProtoBinding struct {
	DID     string    `json:"did"`
	Handle  string    `json:"handle,omitempty"`
	PDS     string    `json:"pds,omitempty"`
	Status  string    `json:"status"`
	BoundAt time.Time `json:"boundAt"`
}

type Device struct {
	ID                string     `json:"id"`
	Name              string     `json:"name"`
	Platform          string     `json:"platform"`
	PublicKeySPKI     string     `json:"publicKeySpki"`
	Status            string     `json:"status"`
	AttestationStatus string     `json:"attestationStatus"`
	EnrolledAt        time.Time  `json:"enrolledAt"`
	RevokedAt         *time.Time `json:"revokedAt,omitempty"`
}

type Account struct {
	ID          string              `json:"id"`
	UserHandle  string              `json:"userHandle"`
	DisplayName string              `json:"displayName"`
	CreatedAt   time.Time           `json:"createdAt"`
	Credentials []PasskeyCredential `json:"credentials"`
	ATProto     *ATProtoBinding     `json:"atproto,omitempty"`
	Devices     []Device            `json:"devices"`
}

type Session struct {
	TokenHash string    `json:"tokenHash"`
	AccountID string    `json:"accountId"`
	CreatedAt time.Time `json:"createdAt"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type State struct {
	SchemaVersion int       `json:"schemaVersion"`
	Accounts      []Account `json:"accounts"`
	Sessions      []Session `json:"sessions"`
}

type AccountView struct {
	ID           string          `json:"id"`
	DisplayName  string          `json:"displayName"`
	CreatedAt    time.Time       `json:"createdAt"`
	PasskeyCount int             `json:"passkeyCount"`
	ATProto      *ATProtoBinding `json:"atproto,omitempty"`
	Devices      []Device        `json:"devices"`
}
