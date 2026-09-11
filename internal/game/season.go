package game

import "time"

type SeasonPhase string

const (
	SeasonPreseason SeasonPhase = "preseason"
	SeasonActive    SeasonPhase = "active"
	SeasonTurnIn    SeasonPhase = "turn-in"
	SeasonArchived  SeasonPhase = "archived"
)

type SeasonConfig struct {
	ID        string
	Name      string
	Location  string
	Chapter   int
	Duration  time.Duration
	TurnInFor time.Duration
}

type SeasonArchive struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Location    string    `json:"location"`
	Chapter     int       `json:"chapter"`
	Earnings    int64     `json:"earnings"`
	TurnIns     int       `json:"turnIns"`
	CompletedAt time.Time `json:"completedAt"`
	Trophy      string    `json:"trophy"`
	Medal       string    `json:"medal"`
}

type StoryState struct {
	Chapter        int      `json:"chapter"`
	Location       string   `json:"location"`
	Unlocked       []string `json:"unlockedLocations"`
	CurrentBeat    int      `json:"currentBeat"`
	CompletedBeats []string `json:"completedBeats"`
}

type WarrenState struct {
	Theme       string   `json:"theme"`
	Decorations []string `json:"decorations"`
}

type EconomyTelemetry struct {
	LifetimeEarned    int64 `json:"lifetimeEarned"`
	UpgradesPurchased int   `json:"upgradesPurchased"`
	OfflineAccruals   int   `json:"offlineAccruals"`
	OfflineEarned     int64 `json:"offlineEarned"`
	SeasonsCompleted  int   `json:"seasonsCompleted"`
}

type StoryBeat struct {
	ID      string `json:"id"`
	NPC     string `json:"npc"`
	Title   string `json:"title"`
	Message string `json:"message"`
}

var seasonCatalog = []SeasonConfig{
	{ID: "season-1-broken-burrow", Name: "Season 1: Welcome to the Warren", Location: "Broken Burrow", Chapter: 1, Duration: 7 * 24 * time.Hour, TurnInFor: 24 * time.Hour},
	{ID: "season-2-scrap-row", Name: "Season 2: Scrap Row", Location: "Scrap Row", Chapter: 2, Duration: 7 * 24 * time.Hour, TurnInFor: 24 * time.Hour},
	{ID: "season-3-carrot-district", Name: "Season 3: Carrot District", Location: "Carrot District", Chapter: 3, Duration: 7 * 24 * time.Hour, TurnInFor: 24 * time.Hour},
}

var storyBeats = []StoryBeat{
	{ID: "hard-as-nails-orders", NPC: "First Sergeant Hard-as-Nails", Title: "Report for duty", Message: "This burrow is broke, Recruit. Fix that before somebody notices."},
	{ID: "stuffy-pallet", NPC: "Private Stuffy", Title: "Unattended supplies", Message: "I found a pallet of carrots. Nobody was guarding it. That means logistics, right?"},
	{ID: "cashmere-margin", NPC: "Captain Cashmere", Title: "Margins win campaigns", Message: "Revenue is ammunition. Upgrade the operations that compound fastest."},
	{ID: "boomboom-workshop", NPC: "Corporal Boomboom", Title: "Everything needs more fuse", Message: "Wealth funds style. Arena power stays equal. I checked. Mostly."},
	{ID: "champ-challenge", NPC: "Da Champ", Title: "Earn the belt", Message: "Stack all the Bunny Bucks you want. In Warren Wars, you still gotta beat Da Champ."},
	{ID: "flopsy-recovery", NPC: "Doc Flopsy", Title: "Do not monetize concussions", Message: "Take the trophy. Leave the permanent combat buffs on the table."},
}

func seasonDefinition(id string) SeasonConfig {
	for _, season := range seasonCatalog {
		if season.ID == id {
			return season
		}
	}
	return seasonCatalog[0]
}

func nextSeason(current string) SeasonConfig {
	for i, season := range seasonCatalog {
		if season.ID == current && i+1 < len(seasonCatalog) {
			return seasonCatalog[i+1]
		}
	}
	return seasonCatalog[0]
}
