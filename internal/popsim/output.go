package popsim

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
)

func WriteJSON(w io.Writer, reports []Report) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(reports)
}

func WriteCSV(w io.Writer, reports []Report) error {
	writer := csv.NewWriter(w)
	header := []string{
		"model", "scenario", "seed", "days", "committee_trials", "committee_size",
		"accounts", "adversarial_accounts", "attested_accounts", "eligible_accounts",
		"adversarial_weight_share", "weight_gini", "weight_hhi",
		"blocking_capture_probability", "blocking_wilson95_low", "blocking_wilson95_high",
		"finality_capture_probability", "finality_wilson95_low", "finality_wilson95_high",
		"blocking_at_least_once_per_day_probability", "mean_adversarial_seats",
		"quorum_online_probability", "mean_online_seats",
		"adversary_cost_usd", "median_days_to_eligibility", "warnings",
	}
	if err := writer.Write(header); err != nil { return err }
	for _, report := range reports {
		row := []string{
			report.Model,
			report.Scenario.Name,
			strconv.FormatInt(report.Scenario.Seed, 10),
			strconv.Itoa(report.Scenario.Days),
			strconv.Itoa(report.Scenario.CommitteeTrials),
			strconv.Itoa(report.Scenario.Policy.CommitteeSize),
			strconv.Itoa(report.Population.Accounts),
			strconv.Itoa(report.Population.AdversarialAccounts),
			strconv.Itoa(report.Population.AttestedAccounts),
			strconv.Itoa(report.Population.EligibleAccounts),
			formatFloat(report.Population.AdversarialWeightShare),
			formatFloat(report.Population.WeightGini),
			formatFloat(report.Population.WeightHHI),
			formatFloat(report.Capture.BlockingCaptureProbability),
			formatFloat(report.Capture.BlockingCaptureWilson95Low),
			formatFloat(report.Capture.BlockingCaptureWilson95High),
			formatFloat(report.Capture.FinalityCaptureProbability),
			formatFloat(report.Capture.FinalityCaptureWilson95Low),
			formatFloat(report.Capture.FinalityCaptureWilson95High),
			formatFloat(report.Capture.BlockingAtLeastOncePerDayProbability),
			formatFloat(report.Capture.MeanAdversarialSeats),
			formatFloat(report.Liveness.QuorumOnlineProbability),
			formatFloat(report.Liveness.MeanOnlineSeats),
			fmt.Sprintf("%.2f", report.Cost.ConfiguredAdversaryTotalUSD),
			formatFloat(report.Cost.MedianDaysToEligibility),
			strings.Join(report.Warnings, " | "),
		}
		if err := writer.Write(row); err != nil { return err }
	}
	writer.Flush()
	return writer.Error()
}

func formatFloat(v float64) string { return strconv.FormatFloat(v, 'f', 8, 64) }
