package popsim

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

func quickScenario(t *testing.T, name string, seed int64) Scenario {
	t.Helper()
	s, err := BuiltinScenario(name, seed)
	if err != nil { t.Fatal(err) }
	s.CommitteeTrials = 250
	if s.Days > 45 { s.Days = 45 }
	return s
}

func TestSimulationIsDeterministic(t *testing.T) {
	s := quickScenario(t, "phone-farm", 77)
	first, err := Run(s)
	if err != nil { t.Fatal(err) }
	second, err := Run(s)
	if err != nil { t.Fatal(err) }
	if !reflect.DeepEqual(first, second) { t.Fatal("same seed and scenario produced different reports") }
}

func TestUnattestedBotFarmCannotEnterCommittee(t *testing.T) {
	s := quickScenario(t, "unattested-bot-farm", 11)
	report, err := Run(s)
	if err != nil { t.Fatal(err) }
	if report.Population.AdversarialAccounts == 0 { t.Fatal("expected adversarial bot accounts") }
	if report.Population.EligibleAdversarialAccounts != 0 { t.Fatalf("unattested bots became eligible: %d", report.Population.EligibleAdversarialAccounts) }
	if report.Population.AdversarialCommitteeWeight != 0 { t.Fatalf("unattested bot weight=%d", report.Population.AdversarialCommitteeWeight) }
	if report.Capture.BlockingCaptureProbability != 0 || report.Capture.FinalityCaptureProbability != 0 { t.Fatalf("unattested bots captured committee: %+v", report.Capture) }
}

func TestRealPhoneFarmCreatesMoreCaptureRiskThanBotFarm(t *testing.T) {
	phone := quickScenario(t, "phone-farm", 22)
	bots := quickScenario(t, "unattested-bot-farm", 22)
	phoneReport, err := Run(phone)
	if err != nil { t.Fatal(err) }
	botReport, err := Run(bots)
	if err != nil { t.Fatal(err) }
	if phoneReport.Population.AdversarialCommitteeWeight <= botReport.Population.AdversarialCommitteeWeight { t.Fatalf("phone weight=%d bot weight=%d", phoneReport.Population.AdversarialCommitteeWeight, botReport.Population.AdversarialCommitteeWeight) }
	if phoneReport.Capture.MeanAdversarialSeats <= botReport.Capture.MeanAdversarialSeats { t.Fatalf("phone seats=%f bot seats=%f", phoneReport.Capture.MeanAdversarialSeats, botReport.Capture.MeanAdversarialSeats) }
	if phoneReport.Cost.ConfiguredAdversaryTotalUSD <= 0 { t.Fatal("real phone attack should expose a non-zero cost model") }
}

func TestProviderOutageReducesCommitteeLiveness(t *testing.T) {
	baseline := quickScenario(t, "baseline", 33)
	outage := baseline
	outage.Name = "forced-outage"
	outage.AppleAvailability = 0.45
	outage.GoogleAvailability = 0.45
	baseReport, err := Run(baseline)
	if err != nil { t.Fatal(err) }
	outageReport, err := Run(outage)
	if err != nil { t.Fatal(err) }
	if outageReport.Liveness.QuorumOnlineProbability >= baseReport.Liveness.QuorumOnlineProbability { t.Fatalf("outage liveness=%f baseline=%f", outageReport.Liveness.QuorumOnlineProbability, baseReport.Liveness.QuorumOnlineProbability) }
}

func TestDeviceChurnCanRemoveParticipantsFromEligibility(t *testing.T) {
	baseline := quickScenario(t, "baseline", 44)
	churn := baseline
	churn.Name = "forced-churn"
	churn.Cohorts = append([]Cohort(nil), baseline.Cohorts...)
	for i := range churn.Cohorts {
		churn.Cohorts[i].DailyDeviceChurnProbability = 1
		churn.Cohorts[i].ReattestationDelayDays = churn.Days + 1
	}
	baseReport, err := Run(baseline)
	if err != nil { t.Fatal(err) }
	churnReport, err := Run(churn)
	if err != nil { t.Fatal(err) }
	if baseReport.Population.EligibleAccounts == 0 { t.Fatal("baseline should produce eligible participants") }
	if churnReport.Population.EligibleAccounts != 0 { t.Fatalf("forced churn left %d eligible participants", churnReport.Population.EligibleAccounts) }
	if churnReport.Liveness.QuorumOnlineProbability != 0 { t.Fatalf("forced churn liveness=%f", churnReport.Liveness.QuorumOnlineProbability) }
}

func TestAttackSweepGrowsConfiguredCost(t *testing.T) {
	scenarios, err := AttackSweep(5, []int{0, 50, 200})
	if err != nil { t.Fatal(err) }
	var previous float64 = -1
	for _, s := range scenarios {
		s.CommitteeTrials = 50
		s.Days = 20
		report, err := Run(s)
		if err != nil { t.Fatal(err) }
		if report.Cost.ConfiguredAdversaryTotalUSD < previous { t.Fatalf("attack cost decreased: %f after %f", report.Cost.ConfiguredAdversaryTotalUSD, previous) }
		previous = report.Cost.ConfiguredAdversaryTotalUSD
	}
}

func TestReportsEncodeJSONAndCSV(t *testing.T) {
	s := quickScenario(t, "smoke", 9)
	report, err := Run(s)
	if err != nil { t.Fatal(err) }
	var jsonOut bytes.Buffer
	if err := WriteJSON(&jsonOut, []Report{report}); err != nil { t.Fatal(err) }
	if !strings.Contains(jsonOut.String(), `"model": "proof-of-play-sim/0.8"`) { t.Fatalf("missing model in JSON: %s", jsonOut.String()) }
	var csvOut bytes.Buffer
	if err := WriteCSV(&csvOut, []Report{report}); err != nil { t.Fatal(err) }
	if !strings.Contains(csvOut.String(), "blocking_capture_probability") || !strings.Contains(csvOut.String(), "smoke") { t.Fatalf("unexpected CSV: %s", csvOut.String()) }
}

func TestDefaultPolicyTracksExecutableAuthorityConstants(t *testing.T) {
	p := DefaultPolicy()
	if p.MissionAward != 25 || p.MaxAuthority != 1000 || p.EligibilityThreshold != 100 || p.DecayGraceDays != 3 || p.DecayPerDayBasisPoints != 50 || p.NewcomerRampDays != 14 || p.CommitteeSize != 64 { t.Fatalf("unexpected executable defaults: %+v", p) }
	if weightedAward(25, 0) != 25 || weightedAward(25, 1) != 6 || weightedAward(25, 2) != 2 || weightedAward(25, 3) != 1 { t.Fatal("device-weight schedule drifted from v0.7") }
}
