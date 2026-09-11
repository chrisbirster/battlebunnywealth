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
	ErrTurnInLocked      = errors.New("season turn-in target not reached")
)

type businessDefinition struct {
	ID          string
	Name        string
	Description string
	Icon        string
	BaseIncome  int64
	BaseCost    int64
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
	if now == nil {
		now = time.Now
	}
	state, err := store.Load()
	if errors.Is(err, ErrNotFound) {
		state = defaultState(now().UTC())
		if err := store.Save(state); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	if state.SchemaVersion != SchemaVersion {
		return nil, fmt.Errorf("game state schema %d unsupported; want %d", state.SchemaVersion, SchemaVersion)
	}
	return &Service{now: now, store: store, state: state}, nil
}

func defaultState(now time.Time) State {
	return State{
		SchemaVersion: SchemaVersion,
		Player: PlayerProfile{
			Name: "Recruit", Callsign: "New Dig", Fur: "cinnamon", Ears: "upright", Uniform: "field-olive",
			Cosmetics: []string{"Recruit Patch"},
		},
		Wallet: Wallet{BunnyBucks: 750},
		Businesses: []BusinessState{{ID: "carrot-logistics", Level: 1}, {ID: "scrap-salvage", Level: 1}, {ID: "tunnel-tolls", Level: 1}},
		Season: SeasonState{ID: SeasonID},
		LastAccruedAt: now,
	}
}

func (s *Service) Snapshot() (Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	earned := s.accrueLocked()
	if err := s.store.Save(s.state); err != nil {
		return Snapshot{}, err
	}
	return s.snapshotLocked(earned), nil
}

func (s *Service) UpdateProfile(profile PlayerProfile) (Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.accrueLocked()
	profile.Name = strings.TrimSpace(profile.Name)
	profile.Callsign = strings.TrimSpace(profile.Callsign)
	if len(profile.Name) < 2 || len(profile.Name) > 24 || len(profile.Callsign) > 20 || !allowedFur(profile.Fur) || !allowedEars(profile.Ears) || !allowedUniform(profile.Uniform) {
		return Snapshot{}, ErrInvalidProfile
	}
	profile.Cosmetics = append([]string(nil), s.state.Player.Cosmetics...)
	s.state.Player = profile
	if err := s.store.Save(s.state); err != nil {
		return Snapshot{}, err
	}
	return s.snapshotLocked(0), nil
}

func (s *Service) UpgradeBusiness(id string) (Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.accrueLocked()
	idx := s.businessIndexLocked(id)
	if idx < 0 {
		return Snapshot{}, ErrUnknownBusiness
	}
	cost := upgradeCost(definition(id), s.state.Businesses[idx].Level)
	if s.state.Wallet.BunnyBucks < cost {
		return Snapshot{}, ErrInsufficientFunds
	}
	s.state.Wallet.BunnyBucks -= cost
	s.state.Businesses[idx].Level++
	if err := s.store.Save(s.state); err != nil {
		return Snapshot{}, err
	}
	return s.snapshotLocked(0), nil
}

func (s *Service) TurnInSeason() (Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.accrueLocked()
	if s.state.Season.Earnings < SeasonTurnInTarget {
		return Snapshot{}, ErrTurnInLocked
	}
	reward := "Season Zero Pennant"
	if s.state.Season.TurnIns > 0 {
		reward = fmt.Sprintf("Season Zero Service Star %d", s.state.Season.TurnIns+1)
	}
	if !contains(s.state.Player.Cosmetics, reward) {
		s.state.Player.Cosmetics = append(s.state.Player.Cosmetics, reward)
	}
	s.state.Wallet.BunnyBucks = 750
	for i := range s.state.Businesses {
		s.state.Businesses[i].Level = 1
	}
	s.state.Season.Earnings = 0
	s.state.Season.TurnIns++
	s.state.LastAccruedAt = s.now().UTC()
	if err := s.store.Save(s.state); err != nil {
		return Snapshot{}, err
	}
	return s.snapshotLocked(0), nil
}

func (s *Service) AdvanceOnboarding() (Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.accrueLocked()
	if s.state.OnboardingStep < 6 { s.state.OnboardingStep++ }
	if err := s.store.Save(s.state); err != nil { return Snapshot{}, err }
	return s.snapshotLocked(0), nil
}

func (s *Service) Standings() ([]Standing, error) {
	snapshot, err := s.Snapshot()
	if err != nil {
		return nil, err
	}
	return []Standing{{Rank: 1, Name: snapshot.Player.Name, Callsign: snapshot.Player.Callsign, BunnyBucks: snapshot.Season.Earnings}}, nil
}

func (s *Service) accrueLocked() int64 {
	now := s.now().UTC()
	if !now.After(s.state.LastAccruedAt) {
		return 0
	}
	elapsed := now.Sub(s.state.LastAccruedAt)
	if elapsed > OfflineEarningsLimit {
		elapsed = OfflineEarningsLimit
	}
	seconds := int64(elapsed / time.Second)
	if seconds <= 0 {
		return 0
	}
	earned := s.totalIncomeLocked() * seconds
	s.state.Wallet.BunnyBucks += earned
	s.state.Season.Earnings += earned
	// Any time beyond the offline cap is intentionally forfeited.
	s.state.LastAccruedAt = now
	return earned
}

func (s *Service) snapshotLocked(accrued int64) Snapshot {
	businesses := make([]BusinessView, 0, len(s.state.Businesses))
	for _, state := range s.state.Businesses {
		def := definition(state.ID)
		businesses = append(businesses, BusinessView{
			ID: state.ID, Name: def.Name, Description: def.Description, Icon: def.Icon, Level: state.Level,
			IncomePerSecond: s.businessIncomeLocked(state.ID), UpgradeCost: upgradeCost(def, state.Level), Synergy: s.synergyLabelLocked(state.ID),
		})
	}
	return Snapshot{
		SchemaVersion: SchemaVersion, Player: cloneState(s.state).Player, Wallet: s.state.Wallet, Businesses: businesses, Season: s.state.Season,
		IncomePerSecond: s.totalIncomeLocked(), AccruedBunnyBucks: accrued, OfflineCapSeconds: int64(OfflineEarningsLimit / time.Second),
		SeasonTurnInTarget: SeasonTurnInTarget, RankedPowerAffected: false, OnboardingStep: s.state.OnboardingStep,
	}
}

func (s *Service) totalIncomeLocked() int64 {
	var total int64
	for _, business := range s.state.Businesses {
		total += s.businessIncomeLocked(business.ID)
	}
	return total
}

func (s *Service) businessIncomeLocked(id string) int64 {
	idx := s.businessIndexLocked(id)
	if idx < 0 { return 0 }
	base := definition(id).BaseIncome * int64(s.state.Businesses[idx].Level)
	basisPoints := int64(10000)
	if s.levelLocked("tunnel-tolls") >= 5 { basisPoints += 500 }
	if id == "carrot-logistics" && s.levelLocked("scrap-salvage") >= 5 { basisPoints += 1000 }
	if id == "scrap-salvage" && s.levelLocked("carrot-logistics") >= 5 { basisPoints += 1000 }
	return base * basisPoints / 10000
}

func (s *Service) synergyLabelLocked(id string) string {
	bonus := 0
	if s.levelLocked("tunnel-tolls") >= 5 { bonus += 5 }
	if id == "carrot-logistics" && s.levelLocked("scrap-salvage") >= 5 { bonus += 10 }
	if id == "scrap-salvage" && s.levelLocked("carrot-logistics") >= 5 { bonus += 10 }
	if bonus == 0 { return "No active synergy" }
	return fmt.Sprintf("+%d%% synergy", bonus)
}

func (s *Service) businessIndexLocked(id string) int {
	for i := range s.state.Businesses { if s.state.Businesses[i].ID == id { return i } }
	return -1
}
func (s *Service) levelLocked(id string) int { idx := s.businessIndexLocked(id); if idx < 0 { return 0 }; return s.state.Businesses[idx].Level }
func definition(id string) businessDefinition { for _, def := range businessCatalog { if def.ID == id { return def } }; return businessDefinition{ID: id, Name: id} }
func upgradeCost(def businessDefinition, level int) int64 { cost := def.BaseCost; for i := 1; i < level; i++ { cost = cost * 165 / 100 }; return cost }
func contains(values []string, value string) bool { for _, candidate := range values { if candidate == value { return true } }; return false }
func allowedFur(value string) bool { return value == "cinnamon" || value == "snow" || value == "charcoal" || value == "spotted" }
func allowedEars(value string) bool { return value == "upright" || value == "lop" || value == "battle-worn" }
func allowedUniform(value string) bool { return value == "field-olive" || value == "desert-tan" || value == "night-black" || value == "medic-white" }
