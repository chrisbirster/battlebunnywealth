package popsim

import "github.com/chrisbirster/battlebunnywealth/internal/proofofplay"

const ModelVersion = "proof-of-play-sim/0.8"

type Policy struct {
	MissionAward           int64 `json:"missionAward"`
	MaxAuthority           int64 `json:"maxAuthority"`
	MaxMissionsPerDay      int   `json:"maxMissionsPerDay"`
	EligibilityThreshold   int64 `json:"eligibilityThreshold"`
	DecayGraceDays         int   `json:"decayGraceDays"`
	DecayPerDayBasisPoints int64 `json:"decayPerDayBasisPoints"`
	NewcomerRampDays       int   `json:"newcomerRampDays"`
	CommitteeSize          int   `json:"committeeSize"`
	QuorumNumerator        int   `json:"quorumNumerator"`
	QuorumDenominator      int   `json:"quorumDenominator"`
}

func DefaultPolicy() Policy {
	authority := proofofplay.DefaultAuthorityConfig()
	protocol := proofofplay.DefaultConfig()
	return Policy{
		MissionAward:           authority.MissionAward,
		MaxAuthority:           authority.MaxAuthority,
		MaxMissionsPerDay:      authority.MaxMissionsPerDay,
		EligibilityThreshold:   authority.EligibilityThreshold,
		DecayGraceDays:         int(authority.DecayGrace.Hours() / 24),
		DecayPerDayBasisPoints: authority.DecayPerDayBPS,
		NewcomerRampDays:       int(authority.NewcomerRamp.Hours() / 24),
		CommitteeSize:          protocol.CommitteeSize,
		QuorumNumerator:        protocol.QuorumNumerator,
		QuorumDenominator:      protocol.QuorumDenominator,
	}
}

type Cohort struct {
	Name                         string  `json:"name"`
	Accounts                     int     `json:"accounts"`
	Adversarial                  bool    `json:"adversarial"`
	StartDay                     int     `json:"startDay"`
	DevicesPerAccount            int     `json:"devicesPerAccount"`
	AttestedFraction             float64 `json:"attestedFraction"`
	DailyActiveProbability       float64 `json:"dailyActiveProbability"`
	MissionCompletionProbability float64 `json:"missionCompletionProbability"`
	CommitteeOnlineProbability   float64 `json:"committeeOnlineProbability"`
	DailyDeviceChurnProbability  float64 `json:"dailyDeviceChurnProbability"`
	ReattestationDelayDays       int     `json:"reattestationDelayDays"`
	Provider                     string  `json:"provider"`
	DeviceCostUSD                float64 `json:"deviceCostUsd"`
	AccountCostUSD               float64 `json:"accountCostUsd"`
	MonthlyOperatingCostUSD      float64 `json:"monthlyOperatingCostUsd"`
}

type Scenario struct {
	Name               string   `json:"name"`
	Seed               int64    `json:"seed"`
	Days               int      `json:"days"`
	CommitteeTrials    int      `json:"committeeTrials"`
	EpochsPerDay       int      `json:"epochsPerDay"`
	AppleAvailability  float64  `json:"appleAvailability"`
	GoogleAvailability float64  `json:"googleAvailability"`
	Policy             Policy   `json:"policy"`
	Cohorts            []Cohort `json:"cohorts"`
}

type PopulationMetrics struct {
	Accounts                    int     `json:"accounts"`
	AdversarialAccounts         int     `json:"adversarialAccounts"`
	AttestedAccounts            int     `json:"attestedAccounts"`
	EligibleAccounts            int     `json:"eligibleAccounts"`
	EligibleAdversarialAccounts int     `json:"eligibleAdversarialAccounts"`
	TotalDevices                int     `json:"totalDevices"`
	AdversarialDevices          int     `json:"adversarialDevices"`
	TotalCommitteeWeight        int64   `json:"totalCommitteeWeight"`
	AdversarialCommitteeWeight  int64   `json:"adversarialCommitteeWeight"`
	AdversarialWeightShare      float64 `json:"adversarialWeightShare"`
	MeanHonestAuthority         float64 `json:"meanHonestAuthority"`
	MeanAdversarialAuthority    float64 `json:"meanAdversarialAuthority"`
	WeightGini                  float64 `json:"weightGini"`
	WeightHHI                   float64 `json:"weightHhi"`
}

type CaptureMetrics struct {
	BlockingThresholdSeats               int     `json:"blockingThresholdSeats"`
	FinalityThresholdSeats               int     `json:"finalityThresholdSeats"`
	BlockingCaptureProbability           float64 `json:"blockingCaptureProbability"`
	BlockingCaptureWilson95Low           float64 `json:"blockingCaptureWilson95Low"`
	BlockingCaptureWilson95High          float64 `json:"blockingCaptureWilson95High"`
	FinalityCaptureProbability           float64 `json:"finalityCaptureProbability"`
	FinalityCaptureWilson95Low           float64 `json:"finalityCaptureWilson95Low"`
	FinalityCaptureWilson95High          float64 `json:"finalityCaptureWilson95High"`
	BlockingAtLeastOncePerDayProbability float64 `json:"blockingAtLeastOncePerDayProbability"`
	FinalityAtLeastOncePerDayProbability float64 `json:"finalityAtLeastOncePerDayProbability"`
	MaxBlockingCaptureStreak             int     `json:"maxBlockingCaptureStreak"`
	MaxFinalityCaptureStreak             int     `json:"maxFinalityCaptureStreak"`
	MeanAdversarialSeats                 float64 `json:"meanAdversarialSeats"`
}

type LivenessMetrics struct {
	MeanOnlineSeats         float64 `json:"meanOnlineSeats"`
	QuorumOnlineProbability float64 `json:"quorumOnlineProbability"`
}

type CostMetrics struct {
	ConfiguredAdversaryHardwareUSD  float64 `json:"configuredAdversaryHardwareUsd"`
	ConfiguredAdversaryAccountsUSD  float64 `json:"configuredAdversaryAccountsUsd"`
	ConfiguredAdversaryOperatingUSD float64 `json:"configuredAdversaryOperatingUsd"`
	ConfiguredAdversaryTotalUSD     float64 `json:"configuredAdversaryTotalUsd"`
	MedianDaysToEligibility         float64 `json:"medianDaysToEligibility"`
}

type Report struct {
	Model      string            `json:"model"`
	Scenario   Scenario          `json:"scenario"`
	Population PopulationMetrics `json:"population"`
	Capture    CaptureMetrics    `json:"capture"`
	Liveness   LivenessMetrics   `json:"liveness"`
	Cost       CostMetrics       `json:"cost"`
	Warnings   []string          `json:"warnings"`
}
