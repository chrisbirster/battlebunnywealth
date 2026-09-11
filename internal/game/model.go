package game

import "time"

const SchemaVersion = 1

const (
	SeasonID             = "season-0-alpha"
	SeasonTurnInTarget   = int64(25000)
	OfflineEarningsLimit = 8 * time.Hour
)

type PlayerProfile struct {
	Name      string   `json:"name"`
	Callsign  string   `json:"callsign"`
	Fur       string   `json:"fur"`
	Ears      string   `json:"ears"`
	Uniform   string   `json:"uniform"`
	Cosmetics []string `json:"cosmetics"`
}

type Wallet struct {
	BunnyBucks int64 `json:"bunnyBucks"`
}

type BusinessState struct {
	ID    string `json:"id"`
	Level int    `json:"level"`
}

type SeasonState struct {
	ID       string `json:"id"`
	Earnings int64  `json:"earnings"`
	TurnIns  int    `json:"turnIns"`
}

type State struct {
	SchemaVersion int             `json:"schemaVersion"`
	Player        PlayerProfile   `json:"player"`
	Wallet        Wallet          `json:"wallet"`
	Businesses    []BusinessState `json:"businesses"`
	Season        SeasonState     `json:"season"`
	LastAccruedAt time.Time       `json:"lastAccruedAt"`
	OnboardingStep int             `json:"onboardingStep"`
}

type BusinessView struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	Icon            string `json:"icon"`
	Level           int    `json:"level"`
	IncomePerSecond int64  `json:"incomePerSecond"`
	UpgradeCost     int64  `json:"upgradeCost"`
	Synergy         string `json:"synergy"`
}

type Snapshot struct {
	SchemaVersion       int            `json:"schemaVersion"`
	Player              PlayerProfile  `json:"player"`
	Wallet              Wallet         `json:"wallet"`
	Businesses          []BusinessView `json:"businesses"`
	Season              SeasonState    `json:"season"`
	IncomePerSecond     int64          `json:"incomePerSecond"`
	AccruedBunnyBucks   int64          `json:"accruedBunnyBucks"`
	OfflineCapSeconds   int64          `json:"offlineCapSeconds"`
	SeasonTurnInTarget  int64          `json:"seasonTurnInTarget"`
	RankedPowerAffected bool           `json:"rankedPowerAffected"`
	OnboardingStep      int            `json:"onboardingStep"`
}

type Standing struct {
	Rank       int    `json:"rank"`
	Name       string `json:"name"`
	Callsign   string `json:"callsign"`
	BunnyBucks int64  `json:"bunnyBucks"`
}
