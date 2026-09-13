package proofofplay

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"strconv"
	"testing"
	"time"
)

func TestAppleAppAttestAssertionValidator(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	spkiDER, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	clientData := []byte("bbw-app-attest-assertion/v1\nchallenge=abc\ndevice=device\npurpose=committee-vote")
	clientHash := sha256.Sum256(clientData)
	appID := "TEAM123456.com.example.bbw"
	rp := sha256.Sum256([]byte(appID))
	authData := make([]byte, 37)
	copy(authData[:32], rp[:])
	authData[32] = 0x01
	binary.BigEndian.PutUint32(authData[33:37], 1)
	nonceInput := append(append([]byte(nil), authData...), clientHash[:]...)
	nonce := sha256.Sum256(nonceInput)
	signature, err := ecdsa.SignASN1(rand.Reader, key, nonce[:])
	if err != nil {
		t.Fatal(err)
	}
	assertion := cborMap([]cborPair{
		{cborText("signature"), cborBytes(signature)},
		{cborText("authenticatorData"), cborBytes(authData)},
	})
	validator := &AppleAppAttestCertificateValidator{}
	result, err := validator.ValidateAppAttestAssertion(context.Background(), AppleAssertionValidationRequest{
		ClientDataHash:        base64.RawURLEncoding.EncodeToString(clientHash[:]),
		AssertionObject:       base64.StdEncoding.EncodeToString(assertion),
		ProviderPublicKeySPKI: base64.RawURLEncoding.EncodeToString(spkiDER),
		BundleID:              "com.example.bbw",
		TeamID:                "TEAM123456",
		PreviousCounter:       0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Counter != 1 {
		t.Fatalf("counter=%d want 1", result.Counter)
	}

	_, err = validator.ValidateAppAttestAssertion(context.Background(), AppleAssertionValidationRequest{
		ClientDataHash:        base64.RawURLEncoding.EncodeToString(clientHash[:]),
		AssertionObject:       base64.StdEncoding.EncodeToString(assertion),
		ProviderPublicKeySPKI: base64.RawURLEncoding.EncodeToString(spkiDER),
		BundleID:              "com.example.bbw",
		TeamID:                "TEAM123456",
		PreviousCounter:       1,
	})
	if err == nil {
		t.Fatal("expected non-increasing counter rejection")
	}
}

func TestAppleAppAttestAssertionRejectsWrongAppID(t *testing.T) {
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	spkiDER, _ := x509.MarshalPKIXPublicKey(&key.PublicKey)
	clientHash := sha256.Sum256([]byte("client-data"))
	rp := sha256.Sum256([]byte("TEAM123456.com.example.bbw"))
	authData := make([]byte, 37)
	copy(authData[:32], rp[:])
	binary.BigEndian.PutUint32(authData[33:37], 1)
	nonce := sha256.Sum256(append(append([]byte(nil), authData...), clientHash[:]...))
	signature, _ := ecdsa.SignASN1(rand.Reader, key, nonce[:])
	assertion := cborMap([]cborPair{
		{cborText("signature"), cborBytes(signature)},
		{cborText("authenticatorData"), cborBytes(authData)},
	})
	validator := &AppleAppAttestCertificateValidator{}
	_, err := validator.ValidateAppAttestAssertion(context.Background(), AppleAssertionValidationRequest{
		ClientDataHash:        base64.RawURLEncoding.EncodeToString(clientHash[:]),
		AssertionObject:       base64.StdEncoding.EncodeToString(assertion),
		ProviderPublicKeySPKI: base64.RawURLEncoding.EncodeToString(spkiDER),
		BundleID:              "com.other.app",
		TeamID:                "TEAM123456",
	})
	if err == nil {
		t.Fatal("expected RP ID rejection")
	}
}

type fakeAppleAssertionProvider struct {
	publicKey string
}

func (f fakeAppleAssertionProvider) Provider() string { return ProviderAppleAppAttest }
func (f fakeAppleAssertionProvider) VerifyEnrollment(context.Context, AttestationChallenge, AttestationEvidence) (AttestationResult, error) {
	return AttestationResult{
		Provider: ProviderAppleAppAttest, ProviderKeyID: "apple-key", ProviderPublicKeySPKI: f.publicKey,
		EvidenceDigest: "digest", IntegrityLabels: []string{"APP_ATTEST_VERIFIED"}, HardwareBacked: true, ProductionEligible: true,
	}, nil
}
func (f fakeAppleAssertionProvider) VerifyAssertion(_ context.Context, _ AttestationChallenge, _ DeviceAttestationRecord, evidence AttestationEvidence) (uint64, error) {
	counter, err := strconv.ParseUint(evidence.Payload, 10, 64)
	if err != nil {
		return 0, err
	}
	return counter, nil
}

func TestAttestationServiceAssertionChallengeAndCounter(t *testing.T) {
	now := time.Date(2026, 9, 12, 20, 0, 0, 0, time.UTC)
	clock := func() time.Time { return now }
	provider := fakeAppleAssertionProvider{publicKey: testDeviceSPKI(t)}
	svc, err := NewAttestationService(clock, NewMemoryAttestationStore(), DefaultAttestationConfig(), provider)
	if err != nil {
		t.Fatal(err)
	}
	deviceKey := testDeviceSPKI(t)
	enrollment, err := svc.Begin("acct", "iphone", deviceKey, ProviderAppleAppAttest)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Complete(context.Background(), "acct", "iphone", AttestationEvidence{
		Provider: ProviderAppleAppAttest, ChallengeID: enrollment.ID, DeviceID: "iphone", Payload: "attestation", KeyID: "apple-key",
	}); err != nil {
		t.Fatal(err)
	}
	challenge, err := svc.BeginAssertion("acct", "iphone", "committee-vote")
	if err != nil {
		t.Fatal(err)
	}
	if challenge.Purpose != AttestationPurposeAssertion || challenge.AssertionPurpose != "committee-vote" || challenge.BindingPayload == "" {
		t.Fatalf("challenge=%+v", challenge)
	}
	record, err := svc.CompleteAssertion(context.Background(), "acct", "iphone", AttestationEvidence{
		Provider: ProviderAppleAppAttest, ChallengeID: challenge.ID, DeviceID: "iphone", Payload: "1", KeyID: "apple-key",
	})
	if err != nil {
		t.Fatal(err)
	}
	if record.AssertionCounter != 1 {
		t.Fatalf("counter=%d want 1", record.AssertionCounter)
	}
	if _, err := svc.CompleteAssertion(context.Background(), "acct", "iphone", AttestationEvidence{
		Provider: ProviderAppleAppAttest, ChallengeID: challenge.ID, DeviceID: "iphone", Payload: "2", KeyID: "apple-key",
	}); !errors.Is(err, ErrAttestationChallengeUsed) {
		t.Fatalf("replay err=%v", err)
	}
	challenge2, err := svc.BeginAssertion("acct", "iphone", "committee-vote")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CompleteAssertion(context.Background(), "acct", "iphone", AttestationEvidence{
		Provider: ProviderAppleAppAttest, ChallengeID: challenge2.ID, DeviceID: "iphone", Payload: "1", KeyID: "apple-key",
	}); !errors.Is(err, ErrAttestationRejected) {
		t.Fatalf("counter replay err=%v", err)
	}
}
