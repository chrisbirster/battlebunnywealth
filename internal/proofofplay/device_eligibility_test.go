package proofofplay

import "testing"

func TestNormalizeDeviceEligibilityRequiresOfficialStoreSignals(t *testing.T) {
	android := NormalizeDeviceEligibility(DeviceAttestationRecord{AccountID: "a", DeviceID: "d", Provider: ProviderGooglePlayIntegrity, Status: AttestationStateVerified, IntegrityLabels: []string{"PLAY_RECOGNIZED", "PLAY_LICENSED", "MEETS_STRONG_INTEGRITY"}, HardwareBacked: true, ProductionEligible: true})
	if !android.ConsensusEligible || android.Platform != "android" {
		t.Fatalf("android=%+v", android)
	}
	unlicensed := NormalizeDeviceEligibility(DeviceAttestationRecord{AccountID: "a", DeviceID: "d2", Provider: ProviderGooglePlayIntegrity, Status: AttestationStateVerified, IntegrityLabels: []string{"PLAY_RECOGNIZED", "MEETS_STRONG_INTEGRITY"}, HardwareBacked: true, ProductionEligible: true})
	if unlicensed.ConsensusEligible {
		t.Fatalf("unlicensed device became consensus eligible: %+v", unlicensed)
	}
	apple := NormalizeDeviceEligibility(DeviceAttestationRecord{AccountID: "a", DeviceID: "d3", Provider: ProviderAppleAppAttest, Status: AttestationStateVerified, IntegrityLabels: []string{"APP_ATTEST_VERIFIED"}, HardwareBacked: true, ProductionEligible: true})
	if !apple.ConsensusEligible || apple.Platform != "apple" {
		t.Fatalf("apple=%+v", apple)
	}
}
