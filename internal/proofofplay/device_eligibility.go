package proofofplay

import "sort"

type DeviceEligibility struct {
	AccountID             string   `json:"accountId"`
	DeviceID              string   `json:"deviceId"`
	Provider              string   `json:"provider"`
	Platform              string   `json:"platform"`
	Verified              bool     `json:"verified"`
	RecognizedOfficialApp bool     `json:"recognizedOfficialApp"`
	LicensedStoreInstall  bool     `json:"licensedStoreInstall"`
	HardwareBacked        bool     `json:"hardwareBacked"`
	ConsensusEligible     bool     `json:"consensusEligible"`
	IntegrityLabels       []string `json:"integrityLabels,omitempty"`
	Reasons               []string `json:"reasons,omitempty"`
}

func NormalizeDeviceEligibility(record DeviceAttestationRecord) DeviceEligibility {
	labels := append([]string(nil), record.IntegrityLabels...)
	sort.Strings(labels)
	out := DeviceEligibility{AccountID: record.AccountID, DeviceID: record.DeviceID, Provider: record.Provider, Verified: record.Status == AttestationStateVerified, HardwareBacked: record.HardwareBacked, IntegrityLabels: labels}
	switch record.Provider {
	case ProviderAppleAppAttest:
		out.Platform = "apple"
		out.RecognizedOfficialApp = hasEligibilityLabel(labels, "APP_ATTEST_VERIFIED")
		out.LicensedStoreInstall = out.RecognizedOfficialApp
	case ProviderGooglePlayIntegrity:
		out.Platform = "android"
		out.RecognizedOfficialApp = hasEligibilityLabel(labels, "PLAY_RECOGNIZED")
		out.LicensedStoreInstall = hasEligibilityLabel(labels, "PLAY_LICENSED")
	default:
		out.Platform = "other"
	}
	if !out.Verified {
		out.Reasons = append(out.Reasons, "attestation not verified")
	}
	if !out.RecognizedOfficialApp {
		out.Reasons = append(out.Reasons, "official application identity not verified")
	}
	if !out.HardwareBacked {
		out.Reasons = append(out.Reasons, "hardware-backed integrity requirement not met")
	}
	if record.Provider == ProviderGooglePlayIntegrity && !out.LicensedStoreInstall {
		out.Reasons = append(out.Reasons, "Google Play entitlement not verified")
	}
	out.ConsensusEligible = out.Verified && out.RecognizedOfficialApp && out.HardwareBacked && record.ProductionEligible
	if record.Provider == ProviderGooglePlayIntegrity {
		out.ConsensusEligible = out.ConsensusEligible && out.LicensedStoreInstall
	}
	return out
}

func hasEligibilityLabel(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}
