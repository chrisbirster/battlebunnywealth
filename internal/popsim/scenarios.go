package popsim

import (
	"fmt"

	"github.com/chrisbirster/battlebunnywealth/internal/proofofplay"
)

func ScenarioNames() []string {
	return []string{
		"baseline",
		"inactive-population",
		"multi-device-households",
		"unattested-bot-farm",
		"phone-farm",
		"delayed-phone-farm",
		"provider-outage",
		"device-churn",
		"colluding-third",
		"compromised-validators",
		"smoke",
	}
}

func BuiltinScenario(name string, seed int64) (Scenario, error) {
	base := Scenario{
		Name: name,
		Seed: seed,
		Days: 60,
		CommitteeTrials: 3000,
		EpochsPerDay: 144,
		AppleAvailability: 1,
		GoogleAvailability: 1,
		Policy: DefaultPolicy(),
	}

	honestApple := Cohort{
		Name: "honest-apple",
		Accounts: 500,
		DevicesPerAccount: 1,
		AttestedFraction: 0.98,
		DailyActiveProbability: 0.68,
		MissionCompletionProbability: 0.38,
		CommitteeOnlineProbability: 0.97,
		DailyDeviceChurnProbability: 0.0008,
		ReattestationDelayDays: 2,
		Provider: proofofplay.ProviderAppleAppAttest,
	}
	honestGoogle := honestApple
	honestGoogle.Name = "honest-google"
	honestGoogle.Provider = proofofplay.ProviderGooglePlayIntegrity

	switch name {
	case "baseline":
		base.Cohorts = []Cohort{honestApple, honestGoogle}
	case "inactive-population":
		inactive := honestGoogle
		inactive.Name = "mostly-inactive"
		inactive.Accounts = 500
		inactive.DailyActiveProbability = 0.12
		inactive.MissionCompletionProbability = 0.18
		honestApple.Accounts = 500
		base.Cohorts = []Cohort{honestApple, inactive}
	case "multi-device-households":
		single := honestApple
		single.Accounts = 700
		multi := honestGoogle
		multi.Name = "multi-device-households"
		multi.Accounts = 300
		multi.DevicesPerAccount = 3
		base.Cohorts = []Cohort{single, multi}
	case "unattested-bot-farm":
		bots := Cohort{
			Name: "emulator-bot-farm",
			Accounts: 2500,
			Adversarial: true,
			DevicesPerAccount: 1,
			AttestedFraction: 0,
			DailyActiveProbability: 1,
			MissionCompletionProbability: 1,
			CommitteeOnlineProbability: 1,
			Provider: "unattested",
			DeviceCostUSD: 1,
			AccountCostUSD: 0.10,
			MonthlyOperatingCostUSD: 0.25,
		}
		base.Cohorts = []Cohort{honestApple, honestGoogle, bots}
	case "phone-farm":
		farm := phoneFarm(250, 0)
		base.Cohorts = []Cohort{honestApple, honestGoogle, farm}
	case "delayed-phone-farm":
		farm := phoneFarm(250, 30)
		base.Days = 75
		base.Cohorts = []Cohort{honestApple, honestGoogle, farm}
	case "provider-outage":
		base.AppleAvailability = 0.82
		base.GoogleAvailability = 0.90
		base.Cohorts = []Cohort{honestApple, honestGoogle}
	case "device-churn":
		churnApple := honestApple
		churnGoogle := honestGoogle
		churnApple.DailyDeviceChurnProbability = 0.018
		churnGoogle.DailyDeviceChurnProbability = 0.018
		churnApple.ReattestationDelayDays = 4
		churnGoogle.ReattestationDelayDays = 4
		base.Cohorts = []Cohort{churnApple, churnGoogle}
	case "colluding-third":
		honestApple.Accounts = 450
		honestGoogle.Accounts = 450
		farm := phoneFarm(450, 0)
		farm.Name = "colluding-real-device-farm"
		base.Cohorts = []Cohort{honestApple, honestGoogle, farm}
	case "compromised-validators":
		honestApple.Accounts = 450
		honestGoogle.Accounts = 450
		compromised := Cohort{
			Name: "compromised-existing-validators",
			Accounts: 100,
			Adversarial: true,
			DevicesPerAccount: 1,
			AttestedFraction: 1,
			DailyActiveProbability: 0.75,
			MissionCompletionProbability: 0.55,
			CommitteeOnlineProbability: 0.99,
			Provider: proofofplay.ProviderGooglePlayIntegrity,
			DeviceCostUSD: 0,
			AccountCostUSD: 0,
			MonthlyOperatingCostUSD: 5,
		}
		base.Cohorts = []Cohort{honestApple, honestGoogle, compromised}
	case "smoke":
		base.Days = 14
		base.CommitteeTrials = 200
		base.Policy.CommitteeSize = 16
		honestApple.Accounts = 80
		honestGoogle.Accounts = 80
		base.Cohorts = []Cohort{honestApple, honestGoogle, phoneFarm(20, 0)}
	default:
		return Scenario{}, fmt.Errorf("unknown simulation scenario %q", name)
	}
	return base, nil
}

func phoneFarm(accounts, startDay int) Cohort {
	return Cohort{
		Name: "real-phone-farm",
		Accounts: accounts,
		Adversarial: true,
		StartDay: startDay,
		DevicesPerAccount: 1,
		AttestedFraction: 0.96,
		DailyActiveProbability: 0.99,
		MissionCompletionProbability: 0.96,
		CommitteeOnlineProbability: 0.995,
		DailyDeviceChurnProbability: 0.001,
		ReattestationDelayDays: 1,
		Provider: proofofplay.ProviderGooglePlayIntegrity,
		DeviceCostUSD: 150,
		AccountCostUSD: 2,
		MonthlyOperatingCostUSD: 4,
	}
}

func AttackSweep(seed int64, attackerAccounts []int) ([]Scenario, error) {
	if len(attackerAccounts) == 0 {
		attackerAccounts = []int{0, 50, 100, 150, 200, 250, 350, 500}
	}
	out := make([]Scenario, 0, len(attackerAccounts))
	for i, count := range attackerAccounts {
		s, err := BuiltinScenario("baseline", seed+int64(i))
		if err != nil { return nil, err }
		s.Name = fmt.Sprintf("attack-sweep-%04d-phones", count)
		if count > 0 {
			s.Cohorts = append(s.Cohorts, phoneFarm(count, 0))
		}
		out = append(out, s)
	}
	return out, nil
}
