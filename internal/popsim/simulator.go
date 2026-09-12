package popsim

import (
	"errors"
	"math"
	"math/rand"
	"sort"

	"github.com/chrisbirster/battlebunnywealth/internal/proofofplay"
)

var ErrInvalidScenario = errors.New("invalid Proof-of-Play simulation scenario")

type actor struct {
	cohortIndex               int
	adversarial               bool
	provider                  string
	attested                  bool
	devices                   int
	dailyActiveProbability    float64
	missionCompletionProbability float64
	committeeOnlineProbability float64
	score                     int64
	firstMissionDay           int
	lastMissionDay            int
	eligibilityDay            int
}

type weightedCandidate struct {
	actorIndex int
	key        float64
}

func Run(s Scenario) (Report, error) {
	if err := validateScenario(s); err != nil {
		return Report{}, err
	}
	rng := rand.New(rand.NewSource(s.Seed))
	actors := buildActors(s, rng)

	for day := 0; day < s.Days; day++ {
		appleUp := rng.Float64() < s.AppleAvailability
		googleUp := rng.Float64() < s.GoogleAvailability
		for i := range actors {
			a := &actors[i]
			applyDecay(a, day, s.Policy)
			if rng.Float64() >= a.dailyActiveProbability {
				continue
			}
			if a.attested && !providerUp(a.provider, appleUp, googleUp) {
				continue
			}
			for mission := 0; mission < s.Policy.MaxMissionsPerDay; mission++ {
				if rng.Float64() >= a.missionCompletionProbability {
					continue
				}
				ordinal := 0
				if a.devices > 1 {
					// Real users may perform a mission from whichever active device is in hand.
					// A phone farm with one account/device naturally remains at ordinal zero.
					ordinal = rng.Intn(a.devices)
				}
				award := weightedAward(s.Policy.MissionAward, ordinal)
				if a.firstMissionDay < 0 {
					a.firstMissionDay = day
				}
				a.lastMissionDay = day
				a.score += award
				if a.score > s.Policy.MaxAuthority {
					a.score = s.Policy.MaxAuthority
				}
				if a.eligibilityDay < 0 && a.score >= s.Policy.EligibilityThreshold {
					a.eligibilityDay = day
				}
			}
		}
	}

	population := populationMetrics(actors, s.Policy)
	capture, liveness := committeeMetrics(actors, s, rng)
	cost := costMetrics(actors, s)
	warnings := warningsFor(s, population, capture, liveness)
	return Report{Model: ModelVersion, Scenario: s, Population: population, Capture: capture, Liveness: liveness, Cost: cost, Warnings: warnings}, nil
}

func validateScenario(s Scenario) error {
	if s.Name == "" || s.Days <= 0 || s.CommitteeTrials <= 0 || s.EpochsPerDay <= 0 {
		return ErrInvalidScenario
	}
	p := s.Policy
	if p.MissionAward <= 0 || p.MaxAuthority <= 0 || p.MaxMissionsPerDay <= 0 || p.EligibilityThreshold < 0 || p.NewcomerRampDays <= 0 || p.CommitteeSize <= 0 || p.QuorumNumerator <= 0 || p.QuorumDenominator <= 0 || p.QuorumNumerator > p.QuorumDenominator {
		return ErrInvalidScenario
	}
	if !probability(s.AppleAvailability) || !probability(s.GoogleAvailability) {
		return ErrInvalidScenario
	}
	accounts := 0
	for _, c := range s.Cohorts {
		if c.Name == "" || c.Accounts < 0 || c.DevicesPerAccount <= 0 || !probability(c.AttestedFraction) || !probability(c.DailyActiveProbability) || !probability(c.MissionCompletionProbability) || !probability(c.CommitteeOnlineProbability) {
			return ErrInvalidScenario
		}
		accounts += c.Accounts
	}
	if accounts == 0 {
		return ErrInvalidScenario
	}
	return nil
}

func probability(v float64) bool { return v >= 0 && v <= 1 }

func buildActors(s Scenario, rng *rand.Rand) []actor {
	count := 0
	for _, cohort := range s.Cohorts {
		count += cohort.Accounts
	}
	actors := make([]actor, 0, count)
	for cohortIndex, cohort := range s.Cohorts {
		for i := 0; i < cohort.Accounts; i++ {
			attested := rng.Float64() < cohort.AttestedFraction
			if cohort.Provider == "unattested" || cohort.Provider == proofofplay.ProviderDevelopment {
				attested = false
			}
			actors = append(actors, actor{
				cohortIndex: cohortIndex,
				adversarial: cohort.Adversarial,
				provider: cohort.Provider,
				attested: attested,
				devices: cohort.DevicesPerAccount,
				dailyActiveProbability: cohort.DailyActiveProbability,
				missionCompletionProbability: cohort.MissionCompletionProbability,
				committeeOnlineProbability: cohort.CommitteeOnlineProbability,
				firstMissionDay: -1,
				lastMissionDay: -1,
				eligibilityDay: -1,
			})
		}
	}
	return actors
}

func applyDecay(a *actor, day int, p Policy) {
	if a.score <= 0 || a.lastMissionDay < 0 {
		return
	}
	if day <= a.lastMissionDay+p.DecayGraceDays {
		return
	}
	reduction := a.score * p.DecayPerDayBasisPoints / 10000
	if reduction < 1 {
		reduction = 1
	}
	a.score -= reduction
	if a.score < 0 {
		a.score = 0
	}
}

func weightedAward(base int64, deviceOrdinal int) int64 {
	percent := proofofplay.DeviceWeightPercent(deviceOrdinal)
	award := base * int64(percent) / 100
	if award < 1 {
		return 1
	}
	return award
}

func committeeWeight(a actor, day int, p Policy) int64 {
	if !a.attested || a.score < p.EligibilityThreshold || a.firstMissionDay < 0 {
		return 0
	}
	elapsed := day - a.firstMissionDay
	percent := 10
	if elapsed >= p.NewcomerRampDays {
		percent = 100
	} else if elapsed > 0 {
		percent = 10 + (90*elapsed)/p.NewcomerRampDays
	}
	weight := a.score * int64(percent) / 100
	if weight < 1 {
		weight = 1
	}
	return weight
}

func providerUp(provider string, appleUp, googleUp bool) bool {
	switch provider {
	case proofofplay.ProviderAppleAppAttest:
		return appleUp
	case proofofplay.ProviderGooglePlayIntegrity:
		return googleUp
	case "unattested", proofofplay.ProviderDevelopment:
		return false
	default:
		return true
	}
}

func populationMetrics(actors []actor, p Policy) PopulationMetrics {
	var m PopulationMetrics
	var honestScore, adversarialScore int64
	var honestCount, adversarialCount int
	weights := make([]float64, 0, len(actors))
	day := 0
	for _, a := range actors {
		if a.lastMissionDay > day {
			day = a.lastMissionDay
		}
	}
	for _, a := range actors {
		m.Accounts++
		m.TotalDevices += a.devices
		if a.attested {
			m.AttestedAccounts++
		}
		if a.adversarial {
			m.AdversarialAccounts++
			m.AdversarialDevices += a.devices
			adversarialScore += a.score
			adversarialCount++
		} else {
			honestScore += a.score
			honestCount++
		}
		weight := committeeWeight(a, day, p)
		if weight > 0 {
			m.EligibleAccounts++
			m.TotalCommitteeWeight += weight
			weights = append(weights, float64(weight))
			if a.adversarial {
				m.EligibleAdversarialAccounts++
				m.AdversarialCommitteeWeight += weight
			}
		}
	}
	if m.TotalCommitteeWeight > 0 {
		m.AdversarialWeightShare = float64(m.AdversarialCommitteeWeight) / float64(m.TotalCommitteeWeight)
	}
	if honestCount > 0 {
		m.MeanHonestAuthority = float64(honestScore) / float64(honestCount)
	}
	if adversarialCount > 0 {
		m.MeanAdversarialAuthority = float64(adversarialScore) / float64(adversarialCount)
	}
	m.WeightGini = gini(weights)
	m.WeightHHI = hhi(weights)
	return m
}

func committeeMetrics(actors []actor, s Scenario, rng *rand.Rand) (CaptureMetrics, LivenessMetrics) {
	quorum := ceilDiv(s.Policy.CommitteeSize*s.Policy.QuorumNumerator, s.Policy.QuorumDenominator)
	blocking := s.Policy.CommitteeSize - quorum + 1
	var blockingHits, finalityHits, quorumOnlineHits int
	var totalAdversarialSeats, totalOnlineSeats int
	maxBlockingStreak, blockingStreak := 0, 0
	maxFinalityStreak, finalityStreak := 0, 0
	day := s.Days - 1

	for trial := 0; trial < s.CommitteeTrials; trial++ {
		appleUp := rng.Float64() < s.AppleAvailability
		googleUp := rng.Float64() < s.GoogleAvailability
		candidates := make([]weightedCandidate, 0, len(actors))
		for i, a := range actors {
			weight := committeeWeight(a, day, s.Policy)
			if weight <= 0 || !providerUp(a.provider, appleUp, googleUp) {
				continue
			}
			u := rng.Float64()
			if u == 0 {
				u = math.SmallestNonzeroFloat64
			}
			// Exponential-race weighted sampling without replacement.
			candidates = append(candidates, weightedCandidate{actorIndex: i, key: -math.Log(u) / float64(weight)})
		}
		sort.Slice(candidates, func(i, j int) bool { return candidates[i].key < candidates[j].key })
		seats := s.Policy.CommitteeSize
		if len(candidates) < seats {
			seats = len(candidates)
		}
		adversarialSeats, onlineSeats := 0, 0
		for seat := 0; seat < seats; seat++ {
			a := actors[candidates[seat].actorIndex]
			if a.adversarial {
				adversarialSeats++
			}
			if rng.Float64() < a.committeeOnlineProbability {
				onlineSeats++
			}
		}
		totalAdversarialSeats += adversarialSeats
		totalOnlineSeats += onlineSeats
		if adversarialSeats >= blocking {
			blockingHits++
			blockingStreak++
			if blockingStreak > maxBlockingStreak {
				maxBlockingStreak = blockingStreak
			}
		} else {
			blockingStreak = 0
		}
		if adversarialSeats >= quorum {
			finalityHits++
			finalityStreak++
			if finalityStreak > maxFinalityStreak {
				maxFinalityStreak = finalityStreak
			}
		} else {
			finalityStreak = 0
		}
		if onlineSeats >= quorum {
			quorumOnlineHits++
		}
	}

	trials := float64(s.CommitteeTrials)
	blockingP := float64(blockingHits) / trials
	finalityP := float64(finalityHits) / trials
	blockLow, blockHigh := wilson(blockingHits, s.CommitteeTrials)
	finalLow, finalHigh := wilson(finalityHits, s.CommitteeTrials)
	capture := CaptureMetrics{
		BlockingThresholdSeats: blocking,
		FinalityThresholdSeats: quorum,
		BlockingCaptureProbability: blockingP,
		BlockingCaptureWilson95Low: blockLow,
		BlockingCaptureWilson95High: blockHigh,
		FinalityCaptureProbability: finalityP,
		FinalityCaptureWilson95Low: finalLow,
		FinalityCaptureWilson95High: finalHigh,
		BlockingAtLeastOncePerDayProbability: 1 - math.Pow(1-blockingP, float64(s.EpochsPerDay)),
		FinalityAtLeastOncePerDayProbability: 1 - math.Pow(1-finalityP, float64(s.EpochsPerDay)),
		MaxBlockingCaptureStreak: maxBlockingStreak,
		MaxFinalityCaptureStreak: maxFinalityStreak,
		MeanAdversarialSeats: float64(totalAdversarialSeats) / trials,
	}
	liveness := LivenessMetrics{
		MeanOnlineSeats: float64(totalOnlineSeats) / trials,
		QuorumOnlineProbability: float64(quorumOnlineHits) / trials,
	}
	return capture, liveness
}

func costMetrics(actors []actor, s Scenario) CostMetrics {
	var m CostMetrics
	eligibilityDays := make([]int, 0)
	for cohortIndex, cohort := range s.Cohorts {
		if !cohort.Adversarial {
			continue
		}
		m.ConfiguredAdversaryHardwareUSD += float64(cohort.Accounts*cohort.DevicesPerAccount) * cohort.DeviceCostUSD
		m.ConfiguredAdversaryAccountsUSD += float64(cohort.Accounts) * cohort.AccountCostUSD
		m.ConfiguredAdversaryOperatingUSD += float64(cohort.Accounts) * cohort.MonthlyOperatingCostUSD * float64(s.Days) / 30
		for _, a := range actors {
			if a.cohortIndex == cohortIndex && a.eligibilityDay >= 0 {
				eligibilityDays = append(eligibilityDays, a.eligibilityDay+1)
			}
		}
	}
	m.ConfiguredAdversaryTotalUSD = m.ConfiguredAdversaryHardwareUSD + m.ConfiguredAdversaryAccountsUSD + m.ConfiguredAdversaryOperatingUSD
	if len(eligibilityDays) == 0 {
		m.MedianDaysToEligibility = -1
	} else {
		sort.Ints(eligibilityDays)
		mid := len(eligibilityDays) / 2
		if len(eligibilityDays)%2 == 0 {
			m.MedianDaysToEligibility = float64(eligibilityDays[mid-1]+eligibilityDays[mid]) / 2
		} else {
			m.MedianDaysToEligibility = float64(eligibilityDays[mid])
		}
	}
	return m
}

func warningsFor(s Scenario, p PopulationMetrics, c CaptureMetrics, l LivenessMetrics) []string {
	warnings := []string{}
	if p.EligibleAccounts < s.Policy.CommitteeSize {
		warnings = append(warnings, "eligible population is smaller than the target committee")
	}
	if p.AdversarialWeightShare >= 1.0/3.0 {
		warnings = append(warnings, "adversarial committee weight is at or above one third")
	}
	if c.BlockingCaptureProbability > 0.01 {
		warnings = append(warnings, "blocking committee capture exceeds one percent per draw")
	}
	if c.FinalityCaptureProbability > 0 {
		warnings = append(warnings, "adversary reached finality-control threshold in the sampled committees")
	}
	if l.QuorumOnlineProbability < 0.99 {
		warnings = append(warnings, "committee online-quorum probability is below 99 percent")
	}
	if s.AppleAvailability < 1 || s.GoogleAvailability < 1 {
		warnings = append(warnings, "provider outage policy is modeled as fail-closed eligibility for the affected provider")
	}
	return warnings
}

func ceilDiv(n, d int) int { return (n + d - 1) / d }

func wilson(successes, trials int) (float64, float64) {
	if trials == 0 {
		return 0, 0
	}
	z := 1.959963984540054
	n := float64(trials)
	p := float64(successes) / n
	denom := 1 + z*z/n
	center := (p + z*z/(2*n)) / denom
	half := z * math.Sqrt((p*(1-p)+z*z/(4*n))/n) / denom
	low, high := center-half, center+half
	if low < 0 { low = 0 }
	if high > 1 { high = 1 }
	return low, high
}

func gini(values []float64) float64 {
	if len(values) == 0 { return 0 }
	v := append([]float64(nil), values...)
	sort.Float64s(v)
	var sum, weighted float64
	for i, x := range v {
		sum += x
		weighted += float64(i+1) * x
	}
	if sum == 0 { return 0 }
	n := float64(len(v))
	return (2*weighted)/(n*sum) - (n+1)/n
}

func hhi(values []float64) float64 {
	var sum float64
	for _, v := range values { sum += v }
	if sum == 0 { return 0 }
	var out float64
	for _, v := range values {
		share := v / sum
		out += share * share
	}
	return out
}
