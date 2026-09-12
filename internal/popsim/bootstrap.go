package popsim

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/chrisbirster/battlebunnywealth/internal/testnet"
)

type BootstrapAttackPolicy struct {
	HonestGenesis           int  `json:"honestGenesis"`
	AttackerPhones          int  `json:"attackerPhones"`
	MaturationDays          int  `json:"maturationDays"`
	ActivationWindowDays    int  `json:"activationWindowDays"`
	MaxActivationsPerWindow int  `json:"maxActivationsPerWindow"`
	SimulationDays          int  `json:"simulationDays"`
	CommitteeTarget         int  `json:"committeeTarget"`
	ImmediateActivation     bool `json:"immediateActivation"`
}

type BootstrapDay struct {
	Day               int     `json:"day"`
	ActiveHonest      int     `json:"activeHonest"`
	ActiveAdversarial int     `json:"activeAdversarial"`
	ActiveTotal       int     `json:"activeTotal"`
	CommitteeSize     int     `json:"committeeSize"`
	Quorum            int     `json:"quorum"`
	BlockingThreshold int     `json:"blockingThreshold"`
	FinalityThreshold int     `json:"finalityThreshold"`
	AdversarialShare  float64 `json:"adversarialShare"`
	CanBlock          bool    `json:"canBlock"`
	CanFinalizeAlone  bool    `json:"canFinalizeAlone"`
}

type BootstrapAttackReport struct {
	Model                   string                `json:"model"`
	Policy                  BootstrapAttackPolicy `json:"policy"`
	FirstBlockingDay        *int                  `json:"firstBlockingDay,omitempty"`
	FirstFinalityControlDay *int                  `json:"firstFinalityControlDay,omitempty"`
	MaxAdversarialShare     float64               `json:"maxAdversarialShare"`
	Timeline                []BootstrapDay        `json:"timeline"`
	Warning                 string                `json:"warning"`
}

func DefaultBootstrapAttackPolicy(attackerPhones int) BootstrapAttackPolicy {
	return BootstrapAttackPolicy{HonestGenesis: 4, AttackerPhones: attackerPhones, MaturationDays: 30, ActivationWindowDays: 7, MaxActivationsPerWindow: 1, SimulationDays: 180, CommitteeTarget: 64}
}

func SimulateBootstrapAttack(policy BootstrapAttackPolicy) (BootstrapAttackReport, error) {
	if policy.HonestGenesis <= 0 {
		return BootstrapAttackReport{}, errors.New("honest genesis count must be positive")
	}
	if policy.AttackerPhones < 0 {
		return BootstrapAttackReport{}, errors.New("attacker phone count cannot be negative")
	}
	if policy.SimulationDays <= 0 {
		return BootstrapAttackReport{}, errors.New("simulation days must be positive")
	}
	if policy.CommitteeTarget <= 0 {
		policy.CommitteeTarget = 64
	}
	if !policy.ImmediateActivation {
		if policy.MaturationDays < 0 || policy.ActivationWindowDays <= 0 || policy.MaxActivationsPerWindow <= 0 {
			return BootstrapAttackReport{}, errors.New("invalid controlled activation policy")
		}
	}
	cfg := testnet.DefaultProtocolConfig()
	cfg.CommitteeTarget = policy.CommitteeTarget
	report := BootstrapAttackReport{Model: "proof-of-play-bootstrap-sim/0.9", Policy: policy, Timeline: make([]BootstrapDay, 0, policy.SimulationDays+1), Warning: "Device attestation is not unique-human proof. Rate limiting delays a real-device farm but does not make four genesis participants permanently Sybil-resistant."}
	adversarial := 0
	for day := 0; day <= policy.SimulationDays; day++ {
		if policy.ImmediateActivation && day == 0 {
			adversarial = policy.AttackerPhones
		}
		if !policy.ImmediateActivation && day >= policy.MaturationDays && (day-policy.MaturationDays)%policy.ActivationWindowDays == 0 {
			adversarial += policy.MaxActivationsPerWindow
			if adversarial > policy.AttackerPhones {
				adversarial = policy.AttackerPhones
			}
		}
		total := policy.HonestGenesis + adversarial
		committee := total
		if committee > policy.CommitteeTarget {
			committee = policy.CommitteeTarget
		}
		// Until the active set exceeds the committee target, every active validator is selected. Above that point this bootstrap model reports the proportional expected seat count; v0.8 Monte Carlo remains the capture-probability model.
		adversarialSeats := adversarial
		if total > committee && total > 0 {
			adversarialSeats = (adversarial*committee + total - 1) / total
		}
		quorum := testnet.QuorumFor(cfg, committee)
		blocking := committee - quorum + 1
		share := 0.0
		if total > 0 {
			share = float64(adversarial) / float64(total)
		}
		d := BootstrapDay{Day: day, ActiveHonest: policy.HonestGenesis, ActiveAdversarial: adversarial, ActiveTotal: total, CommitteeSize: committee, Quorum: quorum, BlockingThreshold: blocking, FinalityThreshold: quorum, AdversarialShare: share, CanBlock: adversarialSeats >= blocking, CanFinalizeAlone: adversarialSeats >= quorum}
		if d.CanBlock && report.FirstBlockingDay == nil {
			x := day
			report.FirstBlockingDay = &x
		}
		if d.CanFinalizeAlone && report.FirstFinalityControlDay == nil {
			x := day
			report.FirstFinalityControlDay = &x
		}
		if share > report.MaxAdversarialShare {
			report.MaxAdversarialShare = share
		}
		report.Timeline = append(report.Timeline, d)
	}
	return report, nil
}

func WriteBootstrapJSON(w io.Writer, reports []BootstrapAttackReport) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(reports)
}
func BootstrapGate(attackerPhones int) error {
	immediate := DefaultBootstrapAttackPolicy(attackerPhones)
	immediate.ImmediateActivation = true
	immediateReport, err := SimulateBootstrapAttack(immediate)
	if err != nil {
		return err
	}
	if immediateReport.FirstBlockingDay == nil || *immediateReport.FirstBlockingDay != 0 {
		return fmt.Errorf("expected immediate activation to demonstrate day-zero blocking risk")
	}
	controlled := DefaultBootstrapAttackPolicy(attackerPhones)
	controlledReport, err := SimulateBootstrapAttack(controlled)
	if err != nil {
		return err
	}
	if controlledReport.FirstBlockingDay != nil && *controlledReport.FirstBlockingDay < controlled.MaturationDays {
		return fmt.Errorf("controlled activation allowed blocking before maturation")
	}
	return nil
}
