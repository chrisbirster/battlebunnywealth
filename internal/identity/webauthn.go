package identity

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
)

var (
	ErrInvalidCredential = errors.New("invalid passkey credential")
	ErrCounterRollback    = errors.New("passkey signature counter rollback")
)

type CredentialDescriptor struct {
	Type       string   `json:"type"`
	ID         string   `json:"id"`
	Transports []string `json:"transports,omitempty"`
}

type RegistrationOptions struct {
	Challenge string `json:"challenge"`
	RP        struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"rp"`
	User struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		DisplayName string `json:"displayName"`
	} `json:"user"`
	PubKeyCredParams []struct {
		Type string `json:"type"`
		Alg  int    `json:"alg"`
	} `json:"pubKeyCredParams"`
	Timeout                 int    `json:"timeout"`
	Attestation             string `json:"attestation"`
	AuthenticatorSelection struct {
		ResidentKey        string `json:"residentKey"`
		RequireResidentKey bool   `json:"requireResidentKey"`
		UserVerification   string `json:"userVerification"`
	} `json:"authenticatorSelection"`
	ExcludeCredentials []CredentialDescriptor `json:"excludeCredentials,omitempty"`
}

type LoginOptions struct {
	Challenge        string `json:"challenge"`
	RPID             string `json:"rpId"`
	Timeout          int    `json:"timeout"`
	UserVerification string `json:"userVerification"`
}

type RegistrationCredential struct {
	ID       string `json:"id"`
	RawID    string `json:"rawId"`
	Type     string `json:"type"`
	Response struct {
		ClientDataJSON    string   `json:"clientDataJSON"`
		AttestationObject string   `json:"attestationObject"`
		Transports        []string `json:"transports,omitempty"`
	} `json:"response"`
}

type AssertionCredential struct {
	ID       string `json:"id"`
	RawID    string `json:"rawId"`
	Type     string `json:"type"`
	Response struct {
		ClientDataJSON    string `json:"clientDataJSON"`
		AuthenticatorData string `json:"authenticatorData"`
		Signature         string `json:"signature"`
		UserHandle        string `json:"userHandle,omitempty"`
	} `json:"response"`
}

type clientData struct {
	Type        string `json:"type"`
	Challenge   string `json:"challenge"`
	Origin      string `json:"origin"`
	CrossOrigin bool   `json:"crossOrigin,omitempty"`
}

func verifyRegistration(rpID, origin, challenge string, input RegistrationCredential) (PasskeyCredential, error) {
	if input.Type != "public-key" || input.ID == "" {
		return PasskeyCredential{}, ErrInvalidCredential
	}
	clientRaw, err := decodeB64(input.Response.ClientDataJSON)
	if err != nil {
		return PasskeyCredential{}, ErrInvalidCredential
	}
	if err := verifyClientData(clientRaw, "webauthn.create", origin, challenge); err != nil {
		return PasskeyCredential{}, err
	}
	attestationRaw, err := decodeB64(input.Response.AttestationObject)
	if err != nil {
		return PasskeyCredential{}, ErrInvalidCredential
	}
	decoded, _, err := decodeCBOR(attestationRaw)
	if err != nil {
		return PasskeyCredential{}, fmt.Errorf("%w: attestation CBOR: %v", ErrInvalidCredential, err)
	}
	object, ok := decoded.(map[any]any)
	if !ok {
		return PasskeyCredential{}, ErrInvalidCredential
	}
	format, _ := object["fmt"].(string)
	if format != "none" {
		return PasskeyCredential{}, fmt.Errorf("%w: unsupported attestation format %q", ErrInvalidCredential, format)
	}
	authData, ok := object["authData"].([]byte)
	if !ok {
		return PasskeyCredential{}, ErrInvalidCredential
	}
	credentialID, x, y, signCount, err := parseRegistrationAuthData(rpID, authData)
	if err != nil {
		return PasskeyCredential{}, err
	}
	rawID, err := decodeB64(input.RawID)
	if err != nil || !bytesEqual(rawID, credentialID) {
		return PasskeyCredential{}, ErrInvalidCredential
	}
	return PasskeyCredential{ID: encodeB64(credentialID), PublicKeyX: encodeB64(x), PublicKeyY: encodeB64(y), SignCount: signCount, Transports: append([]string(nil), input.Response.Transports...)}, nil
}

func verifyAssertion(rpID, origin, challenge, expectedUserHandle string, stored PasskeyCredential, input AssertionCredential) (uint32, error) {
	if input.Type != "public-key" || input.ID != stored.ID {
		return 0, ErrInvalidCredential
	}
	rawID, err := decodeB64(input.RawID)
	storedID, err2 := decodeB64(stored.ID)
	if err != nil || err2 != nil || !bytesEqual(rawID, storedID) {
		return 0, ErrInvalidCredential
	}
	clientRaw, err := decodeB64(input.Response.ClientDataJSON)
	if err != nil {
		return 0, ErrInvalidCredential
	}
	if err := verifyClientData(clientRaw, "webauthn.get", origin, challenge); err != nil {
		return 0, err
	}
	authData, err := decodeB64(input.Response.AuthenticatorData)
	if err != nil || len(authData) < 37 {
		return 0, ErrInvalidCredential
	}
	if err := verifyRPAndFlags(rpID, authData); err != nil {
		return 0, err
	}
	newCount := binary.BigEndian.Uint32(authData[33:37])
	if stored.SignCount != 0 && newCount != 0 && newCount <= stored.SignCount {
		return 0, ErrCounterRollback
	}
	if input.Response.UserHandle != "" && expectedUserHandle != "" {
		userHandle, err := decodeB64(input.Response.UserHandle)
		expected, err2 := decodeB64(expectedUserHandle)
		if err != nil || err2 != nil || !bytesEqual(userHandle, expected) {
			return 0, ErrInvalidCredential
		}
	}
	x, err := decodeB64(stored.PublicKeyX)
	if err != nil {
		return 0, ErrInvalidCredential
	}
	y, err := decodeB64(stored.PublicKeyY)
	if err != nil {
		return 0, ErrInvalidCredential
	}
	pub := ecdsa.PublicKey{Curve: elliptic.P256(), X: new(big.Int).SetBytes(x), Y: new(big.Int).SetBytes(y)}
	if !pub.Curve.IsOnCurve(pub.X, pub.Y) {
		return 0, ErrInvalidCredential
	}
	clientHash := sha256.Sum256(clientRaw)
	message := append(append([]byte(nil), authData...), clientHash[:]...)
	digest := sha256.Sum256(message)
	signature, err := decodeB64(input.Response.Signature)
	if err != nil || !ecdsa.VerifyASN1(&pub, digest[:], signature) {
		return 0, ErrInvalidCredential
	}
	return newCount, nil
}

func verifyClientData(raw []byte, wantType, origin, challenge string) error {
	var client clientData
	if err := json.Unmarshal(raw, &client); err != nil {
		return ErrInvalidCredential
	}
	if client.Type != wantType || client.Origin != origin || client.Challenge != challenge || client.CrossOrigin {
		return ErrInvalidCredential
	}
	return nil
}

func parseRegistrationAuthData(rpID string, data []byte) ([]byte, []byte, []byte, uint32, error) {
	if len(data) < 55 {
		return nil, nil, nil, 0, ErrInvalidCredential
	}
	if err := verifyRPAndFlags(rpID, data); err != nil {
		return nil, nil, nil, 0, err
	}
	if data[32]&0x40 == 0 {
		return nil, nil, nil, 0, ErrInvalidCredential
	}
	signCount := binary.BigEndian.Uint32(data[33:37])
	offset := 37 + 16
	if len(data) < offset+2 {
		return nil, nil, nil, 0, ErrInvalidCredential
	}
	idLen := int(binary.BigEndian.Uint16(data[offset : offset+2]))
	offset += 2
	if idLen == 0 || len(data) < offset+idLen {
		return nil, nil, nil, 0, ErrInvalidCredential
	}
	credentialID := append([]byte(nil), data[offset:offset+idLen]...)
	offset += idLen
	decoded, _, err := decodeCBOR(data[offset:])
	if err != nil {
		return nil, nil, nil, 0, ErrInvalidCredential
	}
	key, ok := decoded.(map[any]any)
	if !ok || intValue(key, 1) != 2 || intValue(key, 3) != -7 || intValue(key, -1) != 1 {
		return nil, nil, nil, 0, ErrInvalidCredential
	}
	x, okX := key[int64(-2)].([]byte)
	y, okY := key[int64(-3)].([]byte)
	if !okX || !okY || len(x) != 32 || len(y) != 32 {
		return nil, nil, nil, 0, ErrInvalidCredential
	}
	curve := elliptic.P256()
	if !curve.IsOnCurve(new(big.Int).SetBytes(x), new(big.Int).SetBytes(y)) {
		return nil, nil, nil, 0, ErrInvalidCredential
	}
	return credentialID, append([]byte(nil), x...), append([]byte(nil), y...), signCount, nil
}

func verifyRPAndFlags(rpID string, data []byte) error {
	if len(data) < 37 {
		return ErrInvalidCredential
	}
	expected := sha256.Sum256([]byte(rpID))
	if !bytesEqual(data[:32], expected[:]) || data[32]&0x01 == 0 {
		return ErrInvalidCredential
	}
	return nil
}

func intValue(m map[any]any, key int64) int64 { v, _ := m[key].(int64); return v }
func decodeB64(value string) ([]byte, error) { return base64.RawURLEncoding.DecodeString(value) }
func encodeB64(value []byte) string { return base64.RawURLEncoding.EncodeToString(value) }
func bytesEqual(a, b []byte) bool { if len(a) != len(b) { return false }; var diff byte; for i := range a { diff |= a[i] ^ b[i] }; return diff == 0 }
