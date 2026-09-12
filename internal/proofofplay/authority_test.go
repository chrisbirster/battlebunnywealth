package proofofplay

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"testing"
	"time"
)

type testDeviceKey struct {
	private *ecdsa.PrivateKey
	spki    string
}

func newTestDeviceKey(t *testing.T) testDeviceKey {
	t.Helper()
	private, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil { t.Fatal(err) }
	spki, err := x509.MarshalPKIXPublicKey(&private.PublicKey)
	if err != nil { t.Fatal(err) }
	return testDeviceKey{private: private, spki: base64.RawURLEncoding.EncodeToString(spki)}
}

func signMission(t *testing.T, key *ecdsa.PrivateKey, payload string) string {
	t.Helper()
	digest := sha256.Sum256([]byte(payload))
	r, s, err := ecdsa.Sign(rand.Reader, key, digest[:])
	if err != nil { t.Fatal(err) }
	raw := make([]byte, 64)
	r.FillBytes(raw[:32])
	s.FillBytes(raw[32:])
	return base64.RawURLEncoding.EncodeToString(raw)
}

func completeOne(t *testing.T, service *AuthorityService, accountID, deviceID string, key testDeviceKey, epoch uint64) AuthoritySnapshot {
	t.Helper()
	snapshot, err := service.IssueMission(accountID, deviceID, key.spki, "unattested", "checkpoint-abc", epoch)
	if err != nil { t.Fatal(err) }
	if snapshot.CurrentMission == nil { t.Fatal("expected current mission") }
	mission := snapshot.CurrentMission
	completed, err := service.CompleteMission(accountID, deviceID, key.spki, mission.ID, signMission(t, key.private, mission.SignedPayload))
	if err != nil { t.Fatal(err) }
	return completed
}

func TestMissionsAccrueBoundedAuthorityAndEnforceDailyLimit(t *testing.T) {
	now := time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC)
	service, err := NewAuthorityService(func() time.Time { return now }, NewMemoryAuthorityStore(), DefaultAuthorityConfig())
	if err != nil { t.Fatal(err) }
	key := newTestDeviceKey(t)
	var snapshot AuthoritySnapshot
	for i := 0; i < 4; i++ {
		snapshot = completeOne(t, service, "account-a", "device-a", key, uint64(i+1))
	}
	if snapshot.Authority.Score != 100 { t.Fatalf("score=%d want 100", snapshot.Authority.Score) }
	if !snapshot.Authority.PrototypeEligible { t.Fatal("100 authority should reach prototype threshold") }
	if snapshot.Authority.ProductionEligible { t.Fatal("v0.6 must not grant production eligibility before attestation") }
	if snapshot.DailyMissionsUsed != 4 { t.Fatalf("daily missions=%d want 4", snapshot.DailyMissionsUsed) }
	if _, err := service.IssueMission("account-a", "device-a", key.spki, "unattested", "checkpoint-abc", 9); !errors.Is(err, ErrMissionLimit) {
		t.Fatalf("fifth mission error=%v want ErrMissionLimit", err)
	}
}

func TestMissionSignatureBindingAndReplayDefense(t *testing.T) {
	now := time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC)
	service, err := NewAuthorityService(func() time.Time { return now }, NewMemoryAuthorityStore(), DefaultAuthorityConfig())
	if err != nil { t.Fatal(err) }
	key := newTestDeviceKey(t)
	wrong := newTestDeviceKey(t)
	snapshot, err := service.IssueMission("account-a", "device-a", key.spki, "unattested", "checkpoint-abc", 3)
	if err != nil { t.Fatal(err) }
	mission := snapshot.CurrentMission
	if _, err := service.CompleteMission("account-a", "device-a", key.spki, mission.ID, signMission(t, wrong.private, mission.SignedPayload)); !errors.Is(err, ErrInvalidMissionSignature) {
		t.Fatalf("wrong key error=%v", err)
	}
	signature := signMission(t, key.private, mission.SignedPayload)
	if _, err := service.CompleteMission("other-account", "device-a", key.spki, mission.ID, signature); !errors.Is(err, ErrMissionBinding) {
		t.Fatalf("binding error=%v", err)
	}
	if _, err := service.CompleteMission("account-a", "device-a", key.spki, mission.ID, signature); err != nil { t.Fatal(err) }
	if _, err := service.CompleteMission("account-a", "device-a", key.spki, mission.ID, signature); !errors.Is(err, ErrMissionCompleted) {
		t.Fatalf("replay error=%v want ErrMissionCompleted", err)
	}
}

func TestMissionExpires(t *testing.T) {
	now := time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC)
	service, err := NewAuthorityService(func() time.Time { return now }, NewMemoryAuthorityStore(), DefaultAuthorityConfig())
	if err != nil { t.Fatal(err) }
	key := newTestDeviceKey(t)
	snapshot, err := service.IssueMission("account-a", "device-a", key.spki, "unattested", "checkpoint-abc", 1)
	if err != nil { t.Fatal(err) }
	mission := snapshot.CurrentMission
	now = now.Add(6 * time.Minute)
	if _, err := service.CompleteMission("account-a", "device-a", key.spki, mission.ID, signMission(t, key.private, mission.SignedPayload)); !errors.Is(err, ErrMissionExpired) {
		t.Fatalf("expired error=%v", err)
	}
}

func TestAuthorityDecaysAfterGrace(t *testing.T) {
	now := time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC)
	service, err := NewAuthorityService(func() time.Time { return now }, NewMemoryAuthorityStore(), DefaultAuthorityConfig())
	if err != nil { t.Fatal(err) }
	key := newTestDeviceKey(t)
	for i := 0; i < 4; i++ { completeOne(t, service, "account-a", "device-a", key, uint64(i)) }
	now = now.Add(4 * 24 * time.Hour)
	snapshot, err := service.Snapshot("account-a")
	if err != nil { t.Fatal(err) }
	if snapshot.Authority.Score >= 100 { t.Fatalf("score=%d should decay after grace", snapshot.Authority.Score) }
	if snapshot.Authority.Score <= 0 { t.Fatalf("slow decay should not erase authority: %d", snapshot.Authority.Score) }
}

func TestCommitteeSelectionIsDeterministicAndWeighted(t *testing.T) {
	now := time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC)
	config := DefaultAuthorityConfig()
	config.EligibilityThreshold = 25
	config.DecayGrace = 100 * 24 * time.Hour
	config.NewcomerRamp = 24 * time.Hour
	service, err := NewAuthorityService(func() time.Time { return now }, NewMemoryAuthorityStore(), config)
	if err != nil { t.Fatal(err) }
	for i, account := range []string{"account-a", "account-b", "account-c"} {
		key := newTestDeviceKey(t)
		for j := 0; j <= i; j++ { completeOne(t, service, account, "device-"+account, key, uint64(i*10+j)) }
	}
	now = now.Add(24 * time.Hour)
	first := service.SelectCommittee(7, "seed-from-chain-head", 2)
	second := service.SelectCommittee(7, "seed-from-chain-head", 2)
	if len(first) != 2 || len(second) != 2 { t.Fatalf("committee sizes %d %d", len(first), len(second)) }
	for i := range first {
		if first[i].AccountID != second[i].AccountID { t.Fatalf("committee not deterministic: %+v vs %+v", first, second) }
		if first[i].CommitteeWeight <= 0 { t.Fatalf("invalid weight: %+v", first[i]) }
		if first[i].ProductionEligible { t.Fatal("v0.6 committee preview cannot be production eligible") }
	}
}
