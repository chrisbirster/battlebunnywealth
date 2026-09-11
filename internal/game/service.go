package game

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

var (
	ErrInsufficientFunds = errors.New("insufficient Bunny Bucks")
	ErrUnknownBusiness   = errors.New("unknown business")
	ErrInvalidProfile    = errors.New("invalid player profile")
	ErrTurnInLocked      = errors.New("season turn-in is not available")
	ErrInvalidWarren     = errors.New("invalid warren customization")
	ErrStoryComplete     = errors.New("current story chapter is complete")
)

type businessDefinition struct {
	ID, Name, Description, Icon string
	BaseIncome, BaseCost        int64
}

var businessCatalog = []businessDefinition{
	{ID: "carrot-logistics", Name: "Questionable Carrot Logistics", Description: "Move carrots fast. Ask questions never.", Icon: "🥕", BaseIncome: 10, BaseCost: 500},
	{ID: "scrap-salvage", Name: "Warren Scrap & Salvage", Description: "If it is unattended, it is inventory.", Icon: "🔧", BaseIncome: 4, BaseCost: 350},
	{ID: "tunnel-tolls", Name: "Strategic Tunnel Tolls", Description: "Own the shortcut. Charge everybody else.", Icon: "🚧", BaseIncome: 25, BaseCost: 1200},
}

type Service struct {
	mu    sync.Mutex
	now   func() time.Time
	store Store
	state State
}

func NewService(now func() time.Time, store Store) (*Service, error) {
	if now == nil { now = time.Now }
	state, err := store.Load()
	if errors.Is(err, ErrNotFound) {
		state = defaultState(now().UTC())
		if err := store.Save(state); err != nil { return nil, err }
	} else if err != nil { return nil, err }
	if state.SchemaVersion == 1 {
		state = migrateV1(state, now().UTC())
		if err := store.Save(state); err != nil { return nil, err }
	}
	if state.SchemaVersion != SchemaVersion { return nil, fmt.Errorf("game state schema %d unsupported; want %d", state.SchemaVersion, SchemaVersion) }
	return &Service{now: now, store: store, state: state}, nil
}

func defaultState(now time.Time) State {
	cfg := seasonCatalog[0]
	return State{SchemaVersion: SchemaVersion, Player: PlayerProfile{Name: "Recruit", Callsign: "New Dig", Fur: "cinnamon", Ears: "upright", Uniform: "field-olive", Cosmetics: []string{"Recruit Patch"}}, Wallet: Wallet{BunnyBucks: 750}, Businesses: defaultBusinesses(), Season: newSeason(cfg, now), Story: StoryState{Chapter: 1, Location: cfg.Location, Unlocked: []string{cfg.Location}}, Warren: WarrenState{Theme: "field-camp", Decorations: []string{"Recruit Flag"}}, LastAccruedAt: now}
}

func migrateV1(state State, now time.Time) State {
	cfg := seasonCatalog[0]
	state.SchemaVersion = SchemaVersion
	state.Season.ID = cfg.ID; state.Season.Name = cfg.Name; state.Season.Phase = SeasonActive; state.Season.StartedAt = now; state.Season.EndsAt = now.Add(cfg.Duration); state.Season.TurnInEndsAt = state.Season.EndsAt.Add(cfg.TurnInFor)
	state.Story = StoryState{Chapter: 1, Location: cfg.Location, Unlocked: []string{cfg.Location}, CurrentBeat: min(state.OnboardingStep, len(storyBeats))}
	state.Warren = WarrenState{Theme: "field-camp", Decorations: []string{"Recruit Flag"}}
	state.Telemetry.LifetimeEarned = state.Season.Earnings
	return state
}

func newSeason(cfg SeasonConfig, now time.Time) SeasonState { return SeasonState{ID: cfg.ID, Name: cfg.Name, Phase: SeasonActive, StartedAt: now, EndsAt: now.Add(cfg.Duration), TurnInEndsAt: now.Add(cfg.Duration + cfg.TurnInFor)} }
func defaultBusinesses() []BusinessState { return []BusinessState{{ID: "carrot-logistics", Level: 1}, {ID: "scrap-salvage", Level: 1}, {ID: "tunnel-tolls", Level: 1}} }

func (s *Service) Snapshot() (Snapshot, error) { s.mu.Lock(); defer s.mu.Unlock(); earned := s.accrueLocked(); s.tickSeasonLocked(); if err := s.store.Save(s.state); err != nil { return Snapshot{}, err }; return s.snapshotLocked(earned), nil }
func (s *Service) UpdateProfile(profile PlayerProfile) (Snapshot, error) { s.mu.Lock(); defer s.mu.Unlock(); s.accrueLocked(); s.tickSeasonLocked(); profile.Name = strings.TrimSpace(profile.Name); profile.Callsign = strings.TrimSpace(profile.Callsign); if len(profile.Name) < 2 || len(profile.Name) > 24 || len(profile.Callsign) > 20 || !allowedFur(profile.Fur) || !allowedEars(profile.Ears) || !allowedUniform(profile.Uniform) { return Snapshot{}, ErrInvalidProfile }; profile.Cosmetics = append([]string(nil), s.state.Player.Cosmetics...); s.state.Player = profile; if err := s.store.Save(s.state); err != nil { return Snapshot{}, err }; return s.snapshotLocked(0), nil }
func (s *Service) UpgradeBusiness(id string) (Snapshot, error) { s.mu.Lock(); defer s.mu.Unlock(); s.accrueLocked(); s.tickSeasonLocked(); if s.state.Season.Phase != SeasonActive { return Snapshot{}, ErrTurnInLocked }; idx := s.businessIndexLocked(id); if idx < 0 { return Snapshot{}, ErrUnknownBusiness }; cost := upgradeCost(definition(id), s.state.Businesses[idx].Level); if s.state.Wallet.BunnyBucks < cost { return Snapshot{}, ErrInsufficientFunds }; s.state.Wallet.BunnyBucks -= cost; s.state.Businesses[idx].Level++; s.state.Telemetry.UpgradesPurchased++; if err := s.store.Save(s.state); err != nil { return Snapshot{}, err }; return s.snapshotLocked(0), nil }
func (s *Service) TurnInSeason() (Snapshot, error) { s.mu.Lock(); defer s.mu.Unlock(); s.accrueLocked(); s.tickSeasonLocked(); if s.state.Season.Phase != SeasonTurnIn || s.state.Season.Earnings-s.state.Season.TurnedInBunnyBucks < SeasonTurnInTarget { return Snapshot{}, ErrTurnInLocked }; s.state.Season.TurnedInBunnyBucks += SeasonTurnInTarget; s.state.Season.TurnIns++; reward := fmt.Sprintf("%s Pennant %d", s.state.Season.Name, s.state.Season.TurnIns); s.addCosmeticLocked(reward); s.state.Warren.Decorations = appendUnique(s.state.Warren.Decorations, reward); if err := s.store.Save(s.state); err != nil { return Snapshot{}, err }; return s.snapshotLocked(0), nil }
func (s *Service) AdvanceOnboarding() (Snapshot, error) { s.mu.Lock(); defer s.mu.Unlock(); s.accrueLocked(); if s.state.OnboardingStep < 6 { s.state.OnboardingStep++ }; if err := s.store.Save(s.state); err != nil { return Snapshot{}, err }; return s.snapshotLocked(0), nil }
func (s *Service) AdvanceStory() (Snapshot, error) { s.mu.Lock(); defer s.mu.Unlock(); s.accrueLocked(); s.tickSeasonLocked(); if s.state.Story.CurrentBeat >= len(storyBeats) { return Snapshot{}, ErrStoryComplete }; beat := storyBeats[s.state.Story.CurrentBeat]; s.state.Story.CompletedBeats = appendUnique(s.state.Story.CompletedBeats, beat.ID); s.state.Story.CurrentBeat++; if err := s.store.Save(s.state); err != nil { return Snapshot{}, err }; return s.snapshotLocked(0), nil }
func (s *Service) UpdateWarren(theme string) (Snapshot, error) { s.mu.Lock(); defer s.mu.Unlock(); if theme != "field-camp" && theme != "scrap-yard" && theme != "carrot-command" { return Snapshot{}, ErrInvalidWarren }; s.state.Warren.Theme = theme; if err := s.store.Save(s.state); err != nil { return Snapshot{}, err }; return s.snapshotLocked(0), nil }
func (s *Service) AdvanceSeasonForDevelopment() (Snapshot, error) { s.mu.Lock(); defer s.mu.Unlock(); s.accrueLocked(); now := s.now().UTC(); switch s.state.Season.Phase { case SeasonActive: s.state.Season.Phase = SeasonTurnIn; s.state.Season.EndsAt = now; s.state.Season.TurnInEndsAt = now.Add(time.Hour); case SeasonTurnIn: s.archiveSeasonLocked(now) }; if err := s.store.Save(s.state); err != nil { return Snapshot{}, err }; return s.snapshotLocked(0), nil }
func (s *Service) Standings() ([]Standing, error) { snapshot, err := s.Snapshot(); if err != nil { return nil, err }; return []Standing{{Rank: 1, Name: snapshot.Player.Name, Callsign: snapshot.Player.Callsign, BunnyBucks: snapshot.Season.Earnings}}, nil }

func (s *Service) accrueLocked() int64 { now := s.now().UTC(); if !now.After(s.state.LastAccruedAt) || s.state.Season.Phase != SeasonActive { return 0 }; effective := now; if effective.After(s.state.Season.EndsAt) { effective = s.state.Season.EndsAt }; elapsed := effective.Sub(s.state.LastAccruedAt); if elapsed <= 0 { s.state.LastAccruedAt = now; return 0 }; capped := elapsed; if capped > OfflineEarningsLimit { capped = OfflineEarningsLimit }; seconds := int64(capped / time.Second); earned := s.totalIncomeLocked() * seconds; s.state.Wallet.BunnyBucks += earned; s.state.Season.Earnings += earned; s.state.Telemetry.LifetimeEarned += earned; if elapsed >= 5*time.Minute { s.state.Telemetry.OfflineAccruals++; s.state.Telemetry.OfflineEarned += earned }; s.state.LastAccruedAt = now; return earned }
func (s *Service) tickSeasonLocked() { now := s.now().UTC(); if s.state.Season.Phase == SeasonActive && !now.Before(s.state.Season.EndsAt) { s.state.Season.Phase = SeasonTurnIn }; if s.state.Season.Phase == SeasonTurnIn && !now.Before(s.state.Season.TurnInEndsAt) { s.archiveSeasonLocked(now) } }
func (s *Service) archiveSeasonLocked(now time.Time) { cfg := seasonDefinition(s.state.Season.ID); trophy := cfg.Name + " Trophy"; medal := "Campaign Medal"; archive := SeasonArchive{ID: s.state.Season.ID, Name: s.state.Season.Name, Location: cfg.Location, Chapter: cfg.Chapter, Earnings: s.state.Season.Earnings, TurnIns: s.state.Season.TurnIns, CompletedAt: now, Trophy: trophy, Medal: medal}; s.state.SeasonHistory = append(s.state.SeasonHistory, archive); s.state.Awards = appendUnique(s.state.Awards, trophy, medal); s.addCosmeticLocked(trophy); s.state.Warren.Decorations = appendUnique(s.state.Warren.Decorations, trophy); s.state.Telemetry.SeasonsCompleted++; next := nextSeason(s.state.Season.ID); s.state.Season = newSeason(next, now); s.state.Wallet.BunnyBucks = 750; s.state.Businesses = defaultBusinesses(); s.state.Story.Chapter = next.Chapter; s.state.Story.Location = next.Location; s.state.Story.Unlocked = appendUnique(s.state.Story.Unlocked, next.Location); s.state.LastAccruedAt = now }
func (s *Service) snapshotLocked(accrued int64) Snapshot { businesses := make([]BusinessView, 0, len(s.state.Businesses)); for _, state := range s.state.Businesses { def := definition(state.ID); businesses = append(businesses, BusinessView{ID: state.ID, Name: def.Name, Description: def.Description, Icon: def.Icon, Level: state.Level, IncomePerSecond: s.businessIncomeLocked(state.ID), UpgradeCost: upgradeCost(def, state.Level), Synergy: s.synergyLabelLocked(state.ID)}) }; var beat *StoryBeat; if s.state.Story.CurrentBeat < len(storyBeats) { copy := storyBeats[s.state.Story.CurrentBeat]; beat = &copy }; remaining := int64(0); now := s.now().UTC(); deadline := s.state.Season.EndsAt; if s.state.Season.Phase == SeasonTurnIn { deadline = s.state.Season.TurnInEndsAt }; if deadline.After(now) { remaining = int64(deadline.Sub(now) / time.Second) }; copyState := cloneState(s.state); return Snapshot{SchemaVersion: SchemaVersion, Player: copyState.Player, Wallet: s.state.Wallet, Businesses: businesses, Season: s.state.Season, SeasonHistory: copyState.SeasonHistory, Story: copyState.Story, Warren: copyState.Warren, Awards: copyState.Awards, Telemetry: s.state.Telemetry, CurrentStoryBeat: beat, IncomePerSecond: s.totalIncomeLocked(), AccruedBunnyBucks: accrued, OfflineCapSeconds: int64(OfflineEarningsLimit / time.Second), SeasonTurnInTarget: SeasonTurnInTarget, SeasonSecondsRemaining: remaining, RankedPowerAffected: false, OnboardingStep: s.state.OnboardingStep} }
func (s *Service) totalIncomeLocked() int64 { if s.state.Season.Phase != SeasonActive { return 0 }; var total int64; for _, business := range s.state.Businesses { total += s.businessIncomeLocked(business.ID) }; return total }
func (s *Service) businessIncomeLocked(id string) int64 { idx := s.businessIndexLocked(id); if idx < 0 { return 0 }; base := definition(id).BaseIncome * int64(s.state.Businesses[idx].Level); basis := int64(10000); if s.levelLocked("tunnel-tolls") >= 5 { basis += 500 }; if id == "carrot-logistics" && s.levelLocked("scrap-salvage") >= 5 { basis += 1000 }; if id == "scrap-salvage" && s.levelLocked("carrot-logistics") >= 5 { basis += 1000 }; return base * basis / 10000 }
func (s *Service) synergyLabelLocked(id string) string { bonus := 0; if s.levelLocked("tunnel-tolls") >= 5 { bonus += 5 }; if id == "carrot-logistics" && s.levelLocked("scrap-salvage") >= 5 { bonus += 10 }; if id == "scrap-salvage" && s.levelLocked("carrot-logistics") >= 5 { bonus += 10 }; if bonus == 0 { return "No active synergy" }; return fmt.Sprintf("+%d%% synergy", bonus) }
func (s *Service) businessIndexLocked(id string) int { for i := range s.state.Businesses { if s.state.Businesses[i].ID == id { return i } }; return -1 }
func (s *Service) levelLocked(id string) int { idx := s.businessIndexLocked(id); if idx < 0 { return 0 }; return s.state.Businesses[idx].Level }
func (s *Service) addCosmeticLocked(value string) { s.state.Player.Cosmetics = appendUnique(s.state.Player.Cosmetics, value) }
func definition(id string) businessDefinition { for _, def := range businessCatalog { if def.ID == id { return def } }; return businessDefinition{ID: id, Name: id} }
func upgradeCost(def businessDefinition, level int) int64 { cost := def.BaseCost; for i := 1; i < level; i++ { cost = cost * 165 / 100 }; return cost }
func contains(values []string, value string) bool { for _, candidate := range values { if candidate == value { return true } }; return false }
func appendUnique(values []string, additions ...string) []string { for _, addition := range additions { if !contains(values, addition) { values = append(values, addition) } }; return values }
func allowedFur(value string) bool { return value == "cinnamon" || value == "snow" || value == "charcoal" || value == "spotted" }
func allowedEars(value string) bool { return value == "upright" || value == "lop" || value == "battle-worn" }
func allowedUniform(value string) bool { return value == "field-olive" || value == "desert-tan" || value == "night-black" || value == "medic-white" }
func min(a, b int) int { if a < b { return a }; return b }
