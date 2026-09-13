package proofofplay

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
)

type AppleAssertionValidationRequest struct {
	ClientDataHash        string
	AssertionObject       string
	ProviderPublicKeySPKI string
	BundleID              string
	TeamID                string
	PreviousCounter       uint64
}

type AppleAssertionValidationResult struct {
	Counter uint64
}

type AppleAssertionValidator interface {
	ValidateAppAttestAssertion(context.Context, AppleAssertionValidationRequest) (AppleAssertionValidationResult, error)
}

type AttestationAssertionVerifier interface {
	VerifyAssertion(context.Context, AttestationChallenge, DeviceAttestationRecord, AttestationEvidence) (uint64, error)
}

func (v AppleAppAttestVerifier) VerifyAssertion(ctx context.Context, challenge AttestationChallenge, record DeviceAttestationRecord, evidence AttestationEvidence) (uint64, error) {
	validator, ok := v.Validator.(AppleAssertionValidator)
	if !ok || validator == nil {
		return 0, errors.New("Apple App Attest assertion validator not configured")
	}
	if evidence.Payload == "" || evidence.KeyID == "" {
		return 0, errors.New("missing Apple App Attest assertion or key id")
	}
	if evidence.KeyID != record.ProviderKeyID {
		return 0, errors.New("Apple App Attest assertion key mismatch")
	}
	if record.ProviderPublicKeySPKI == "" {
		return 0, errors.New("Apple App Attest public key is not stored")
	}
	result, err := validator.ValidateAppAttestAssertion(ctx, AppleAssertionValidationRequest{
		ClientDataHash:        challenge.RequestHash,
		AssertionObject:       evidence.Payload,
		ProviderPublicKeySPKI: record.ProviderPublicKeySPKI,
		BundleID:              v.Config.AppleBundleID,
		TeamID:                v.Config.AppleTeamID,
		PreviousCounter:       record.AssertionCounter,
	})
	if err != nil {
		return 0, err
	}
	return result.Counter, nil
}

func (v *AppleAppAttestCertificateValidator) ValidateAppAttestAssertion(_ context.Context, req AppleAssertionValidationRequest) (AppleAssertionValidationResult, error) {
	if req.BundleID == "" || req.TeamID == "" || req.ClientDataHash == "" || req.ProviderPublicKeySPKI == "" {
		return AppleAssertionValidationResult{}, errors.New("incomplete Apple App Attest assertion request")
	}
	objectBytes, err := base64.StdEncoding.DecodeString(req.AssertionObject)
	if err != nil {
		objectBytes, err = base64.RawStdEncoding.DecodeString(req.AssertionObject)
	}
	if err != nil {
		return AppleAssertionValidationResult{}, fmt.Errorf("decode App Attest assertion: %w", err)
	}
	decoded, rest, err := decodeCBOR(objectBytes, 0)
	if err != nil || len(rest) != 0 {
		return AppleAssertionValidationResult{}, fmt.Errorf("decode App Attest assertion CBOR: %w", err)
	}
	top, ok := decoded.(map[any]any)
	if !ok {
		return AppleAssertionValidationResult{}, errors.New("App Attest assertion is not a CBOR map")
	}
	signature, ok := top["signature"].([]byte)
	if !ok || len(signature) == 0 {
		return AppleAssertionValidationResult{}, errors.New("missing App Attest assertion signature")
	}
	authData, ok := top["authenticatorData"].([]byte)
	if !ok || len(authData) < 37 {
		return AppleAssertionValidationResult{}, errors.New("missing or short App Attest assertion authenticator data")
	}
	spkiDER, err := decodeBase64Flexible(req.ProviderPublicKeySPKI)
	if err != nil {
		return AppleAssertionValidationResult{}, errors.New("invalid stored App Attest public key")
	}
	parsed, err := x509.ParsePKIXPublicKey(spkiDER)
	if err != nil {
		return AppleAssertionValidationResult{}, fmt.Errorf("parse stored App Attest public key: %w", err)
	}
	pub, ok := parsed.(*ecdsa.PublicKey)
	if !ok || pub.Curve != elliptic.P256() {
		return AppleAssertionValidationResult{}, errors.New("stored App Attest public key is not P-256")
	}
	clientHash, err := decodeBase64Flexible(req.ClientDataHash)
	if err != nil || len(clientHash) != sha256.Size {
		return AppleAssertionValidationResult{}, errors.New("invalid App Attest assertion clientDataHash")
	}
	nonceInput := make([]byte, 0, len(authData)+len(clientHash))
	nonceInput = append(nonceInput, authData...)
	nonceInput = append(nonceInput, clientHash...)
	nonce := sha256.Sum256(nonceInput)
	if !ecdsa.VerifyASN1(pub, nonce[:], signature) {
		return AppleAssertionValidationResult{}, errors.New("App Attest assertion signature is invalid")
	}
	appID := req.TeamID + "." + req.BundleID
	rp := sha256.Sum256([]byte(appID))
	if !bytes.Equal(authData[:32], rp[:]) {
		return AppleAssertionValidationResult{}, errors.New("App Attest assertion relying-party identifier mismatch")
	}
	counter := uint64(binary.BigEndian.Uint32(authData[33:37]))
	if counter == 0 || counter <= req.PreviousCounter {
		return AppleAssertionValidationResult{}, fmt.Errorf("App Attest assertion counter %d is not greater than %d", counter, req.PreviousCounter)
	}
	return AppleAssertionValidationResult{Counter: counter}, nil
}

func appAttestPublicKeySPKI(attestationObject string) (string, error) {
	objectBytes, err := base64.StdEncoding.DecodeString(attestationObject)
	if err != nil {
		objectBytes, err = base64.RawStdEncoding.DecodeString(attestationObject)
	}
	if err != nil {
		return "", fmt.Errorf("decode App Attest object for public key: %w", err)
	}
	decoded, rest, err := decodeCBOR(objectBytes, 0)
	if err != nil || len(rest) != 0 {
		return "", fmt.Errorf("decode App Attest CBOR for public key: %w", err)
	}
	top, ok := decoded.(map[any]any)
	if !ok {
		return "", errors.New("App Attest object is not a CBOR map")
	}
	stmt, ok := top["attStmt"].(map[any]any)
	if !ok {
		return "", errors.New("missing App Attest statement")
	}
	chainValues, ok := stmt["x5c"].([]any)
	if !ok || len(chainValues) == 0 {
		return "", errors.New("missing App Attest certificate chain")
	}
	leafDER, ok := chainValues[0].([]byte)
	if !ok {
		return "", errors.New("invalid App Attest leaf certificate")
	}
	leaf, err := x509.ParseCertificate(leafDER)
	if err != nil {
		return "", fmt.Errorf("parse App Attest leaf certificate: %w", err)
	}
	pub, ok := leaf.PublicKey.(*ecdsa.PublicKey)
	if !ok || pub.Curve != elliptic.P256() {
		return "", errors.New("App Attest leaf key is not P-256")
	}
	spki, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return "", fmt.Errorf("marshal App Attest public key: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(spki), nil
}

func (s *AttestationService) BeginAssertion(accountID, deviceID, purpose string) (AttestationChallenge, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	purpose = strings.TrimSpace(purpose)
	if purpose == "" || len(purpose) > 128 || strings.ContainsAny(purpose, "\r\n") {
		return AttestationChallenge{}, ErrAttestationBinding
	}
	record, ok := s.deviceLocked(accountID, deviceID)
	if !ok || record.Provider != ProviderAppleAppAttest || record.Status != AttestationStateVerified || record.ProviderPublicKeySPKI == "" {
		return AttestationChallenge{}, ErrAttestationRejected
	}
	provider, ok := s.providers[ProviderAppleAppAttest]
	if !ok {
		return AttestationChallenge{}, ErrAttestationProvider
	}
	if _, ok := provider.(AttestationAssertionVerifier); !ok {
		return AttestationChallenge{}, ErrAttestationProvider
	}
	now := s.now().UTC()
	for i := range s.state.Challenges {
		challenge := &s.state.Challenges[i]
		if challenge.AccountID == accountID && challenge.DeviceID == deviceID && challenge.Provider == ProviderAppleAppAttest && challenge.Purpose == AttestationPurposeAssertion && challenge.AssertionPurpose == purpose && challenge.UsedAt == nil && challenge.ExpiresAt.After(now) {
			return publicAttestationChallenge(*challenge), nil
		}
	}
	id, err := attestationToken(18)
	if err != nil {
		return AttestationChallenge{}, err
	}
	nonce, err := attestationToken(32)
	if err != nil {
		return AttestationChallenge{}, err
	}
	binding := fmt.Sprintf("bbw-app-attest-assertion/v1\nchallenge=%s\ndevice=%s\npurpose=%s", nonce, deviceID, purpose)
	requestHashBytes := sha256.Sum256([]byte(binding))
	challenge := AttestationChallenge{
		ID: id, AccountID: accountID, DeviceID: deviceID, Provider: ProviderAppleAppAttest,
		Purpose: AttestationPurposeAssertion, AssertionPurpose: purpose, Nonce: nonce,
		BindingPayload: binding, RequestHash: base64.RawURLEncoding.EncodeToString(requestHashBytes[:]),
		IssuedAt: now, ExpiresAt: now.Add(s.config.ChallengeTTL),
	}
	s.state.Challenges = append(s.state.Challenges, challenge)
	if err := s.store.Save(s.state); err != nil {
		return AttestationChallenge{}, err
	}
	return publicAttestationChallenge(challenge), nil
}

func (s *AttestationService) CompleteAssertion(ctx context.Context, accountID, deviceID string, evidence AttestationEvidence) (DeviceAttestationRecord, error) {
	s.mu.Lock()
	idx := -1
	for i := range s.state.Challenges {
		if s.state.Challenges[i].ID == evidence.ChallengeID {
			idx = i
			break
		}
	}
	if idx < 0 {
		s.mu.Unlock()
		return DeviceAttestationRecord{}, ErrAttestationChallengeNotFound
	}
	challenge := s.state.Challenges[idx]
	now := s.now().UTC()
	if challenge.UsedAt != nil {
		s.mu.Unlock()
		return DeviceAttestationRecord{}, ErrAttestationChallengeUsed
	}
	if !challenge.ExpiresAt.After(now) {
		s.mu.Unlock()
		return DeviceAttestationRecord{}, ErrAttestationChallengeExpired
	}
	if challenge.Purpose != AttestationPurposeAssertion || challenge.AccountID != accountID || challenge.DeviceID != deviceID || evidence.DeviceID != deviceID || evidence.Provider != ProviderAppleAppAttest {
		s.mu.Unlock()
		return DeviceAttestationRecord{}, ErrAttestationBinding
	}
	record, ok := s.deviceLocked(accountID, deviceID)
	if !ok || record.Provider != ProviderAppleAppAttest || record.Status != AttestationStateVerified {
		s.mu.Unlock()
		return DeviceAttestationRecord{}, ErrAttestationRejected
	}
	provider, ok := s.providers[ProviderAppleAppAttest]
	if !ok {
		s.mu.Unlock()
		return DeviceAttestationRecord{}, ErrAttestationProvider
	}
	assertionVerifier, ok := provider.(AttestationAssertionVerifier)
	if !ok {
		s.mu.Unlock()
		return DeviceAttestationRecord{}, ErrAttestationProvider
	}
	s.state.Challenges[idx].UsedAt = &now
	if err := s.store.Save(s.state); err != nil {
		s.mu.Unlock()
		return DeviceAttestationRecord{}, err
	}
	s.mu.Unlock()

	counter, err := assertionVerifier.VerifyAssertion(ctx, challenge, record, evidence)
	if err != nil {
		return DeviceAttestationRecord{}, fmt.Errorf("%w: %v", ErrAttestationRejected, err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	currentIndex := -1
	for i := range s.state.Devices {
		if s.state.Devices[i].AccountID == accountID && s.state.Devices[i].DeviceID == deviceID {
			currentIndex = i
			break
		}
	}
	if currentIndex < 0 {
		return DeviceAttestationRecord{}, ErrAttestationRejected
	}
	current := s.state.Devices[currentIndex]
	if counter <= current.AssertionCounter {
		return DeviceAttestationRecord{}, fmt.Errorf("%w: App Attest assertion counter did not advance", ErrAttestationRejected)
	}
	current.AssertionCounter = counter
	current.VerifiedAt = s.now().UTC()
	s.state.Devices[currentIndex] = current
	if err := s.store.Save(s.state); err != nil {
		return DeviceAttestationRecord{}, err
	}
	return current, nil
}

func (s *AttestationService) deviceLocked(accountID, deviceID string) (DeviceAttestationRecord, bool) {
	for _, record := range s.state.Devices {
		if record.AccountID == accountID && record.DeviceID == deviceID {
			return record, true
		}
	}
	return DeviceAttestationRecord{}, false
}
