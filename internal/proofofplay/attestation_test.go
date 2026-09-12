package proofofplay

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"testing"
	"time"
)

type fakeAppleValidator struct {
	result AppleValidationResult
	err    error
}

func (f fakeAppleValidator) ValidateAppAttest(context.Context, AppleValidationRequest) (AppleValidationResult, error) {
	return f.result, f.err
}

type fakePlayDecoder struct {
	verdict PlayIntegrityVerdict
	err     error
}

func (f fakePlayDecoder) DecodeIntegrityToken(context.Context, string, string) (PlayIntegrityVerdict, error) {
	return f.verdict, f.err
}

func testDeviceSPKI(t *testing.T) string {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	return base64.RawURLEncoding.EncodeToString(raw)
}

func TestAttestationChallengeSingleUseAndBinding(t *testing.T) {
	now := time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC)
	clock := func() time.Time { return now }
	store := NewMemoryAttestationStore()
	svc, err := NewAttestationService(clock, store, DefaultAttestationConfig(), DevelopmentAttestationVerifier{})
	if err != nil {
		t.Fatal(err)
	}
	key := testDeviceSPKI(t)
	challenge, err := svc.Begin("acct", "dev", key, ProviderDevelopment)
	if err != nil {
		t.Fatal(err)
	}
	if challenge.RequestHash == "" || challenge.DevicePublicKeyHash == "" || challenge.BindingPayload == "" {
		t.Fatal("expected key-bound challenge")
	}
	if _, err := svc.Complete(context.Background(), "acct", "other", AttestationEvidence{Provider: ProviderDevelopment, ChallengeID: challenge.ID, DeviceID: "other", Payload: "allow:" + challenge.Nonce}); !errors.Is(err, ErrAttestationBinding) {
		t.Fatalf("binding err=%v", err)
	}
	record, err := svc.Complete(context.Background(), "acct", "dev", AttestationEvidence{Provider: ProviderDevelopment, ChallengeID: challenge.ID, DeviceID: "dev", Payload: "allow:" + challenge.Nonce})
	if err != nil {
		t.Fatal(err)
	}
	if record.ProductionEligible || record.HardwareBacked {
		t.Fatal("development provider must never be production eligible")
	}
	if _, err := svc.Complete(context.Background(), "acct", "dev", AttestationEvidence{Provider: ProviderDevelopment, ChallengeID: challenge.ID, DeviceID: "dev", Payload: "allow:" + challenge.Nonce}); !errors.Is(err, ErrAttestationChallengeUsed) {
		t.Fatalf("replay err=%v", err)
	}
}

func TestAttestationChallengeExpires(t *testing.T) {
	now := time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC)
	clock := func() time.Time { return now }
	svc, _ := NewAttestationService(clock, NewMemoryAttestationStore(), AttestationConfig{ChallengeTTL: time.Minute}, DevelopmentAttestationVerifier{})
	challenge, _ := svc.Begin("acct", "dev", testDeviceSPKI(t), ProviderDevelopment)
	now = now.Add(2 * time.Minute)
	_, err := svc.Complete(context.Background(), "acct", "dev", AttestationEvidence{Provider: ProviderDevelopment, ChallengeID: challenge.ID, DeviceID: "dev", Payload: "allow:" + challenge.Nonce})
	if !errors.Is(err, ErrAttestationChallengeExpired) {
		t.Fatalf("err=%v", err)
	}
}

func TestAppleAdapterRequiresHardwareBackedResult(t *testing.T) {
	verifier := AppleAppAttestVerifier{Config: AttestationConfig{AppleBundleID: "com.example.bbw", AppleTeamID: "TEAM", AppleEnvironment: "development"}, Validator: fakeAppleValidator{result: AppleValidationResult{KeyID: "key", HardwareBacked: true}}}
	result, err := verifier.VerifyEnrollment(context.Background(), AttestationChallenge{Nonce: "nonce", RequestHash: "hash"}, AttestationEvidence{KeyID: "key", Payload: "object"})
	if err != nil {
		t.Fatal(err)
	}
	if !result.ProductionEligible || !result.HardwareBacked {
		t.Fatalf("result=%+v", result)
	}
}

func TestPlayIntegrityAdapterPolicy(t *testing.T) {
	challenge := AttestationChallenge{ID: "challenge", DeviceID: "device", Nonce: "nonce", RequestHash: "hash"}
	verifier := GooglePlayIntegrityVerifier{Config: AttestationConfig{AndroidPackageName: "com.example.bbw", AndroidAllowedCertificates: []string{"cert"}, AndroidRequireStrongIntegrity: true}, Decoder: fakePlayDecoder{verdict: PlayIntegrityVerdict{RequestHash: "hash", PackageName: "com.example.bbw", CertificateDigests: []string{"cert"}, AppRecognitionVerdict: "PLAY_RECOGNIZED", AppLicensingVerdict: "LICENSED", DeviceIntegrity: []string{"MEETS_DEVICE_INTEGRITY", "MEETS_STRONG_INTEGRITY"}}}}
	result, err := verifier.VerifyEnrollment(context.Background(), challenge, AttestationEvidence{Payload: "token"})
	if err != nil {
		t.Fatal(err)
	}
	if !result.ProductionEligible || !result.HardwareBacked {
		t.Fatalf("result=%+v", result)
	}
}

func TestPlayIntegrityRejectsWrongRequestHash(t *testing.T) {
	verifier := GooglePlayIntegrityVerifier{Config: AttestationConfig{AndroidPackageName: "pkg"}, Decoder: fakePlayDecoder{verdict: PlayIntegrityVerdict{RequestHash: "wrong", PackageName: "pkg", AppRecognitionVerdict: "PLAY_RECOGNIZED", DeviceIntegrity: []string{"MEETS_DEVICE_INTEGRITY"}}}}
	_, err := verifier.VerifyEnrollment(context.Background(), AttestationChallenge{ID: "c", DeviceID: "d", Nonce: "n", RequestHash: "right"}, AttestationEvidence{Payload: "token"})
	if err == nil {
		t.Fatal("expected rejection")
	}
}

func TestDeviceWeights(t *testing.T) {
	want := []int{100, 25, 10, 2, 2}
	for i, w := range want {
		if got := DeviceWeightPercent(i); got != w {
			t.Fatalf("ordinal %d=%d want %d", i, got, w)
		}
	}
}
