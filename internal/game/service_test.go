package game

import (
	"path/filepath"
	"testing"
	"time"
)

func TestOfflineEarningsAreAutomaticAndCapped(t *testing.T) {
	now := time.Date(2026, 9, 11, 12, 0, 0, 0, time.UTC)
	clock := func() time.Time { return now }
	service, err := NewService(clock, NewMemoryStore())
	if err != nil { t.Fatal(err) }
	initial, err := service.Snapshot()
	if err != nil { t.Fatal(err) }
	if initial.IncomePerSecond != 39 { t.Fatalf("income=%d want 39", initial.IncomePerSecond) }
	now = now.Add(12 * time.Hour)
	after, err := service.Snapshot()
	if err != nil { t.Fatal(err) }
	want := int64(39 * 8 * 60 * 60)
	if after.AccruedBunnyBucks != want { t.Fatalf("accrued=%d want %d", after.AccruedBunnyBucks, want) }
	if after.Wallet.BunnyBucks != 750+want { t.Fatalf("wallet=%d want %d", after.Wallet.BunnyBucks, 750+want) }
}

func TestUpgradeAndSynergy(t *testing.T) {
	now := time.Date(2026, 9, 11, 12, 0, 0, 0, time.UTC)
	service, err := NewService(func() time.Time { return now }, NewMemoryStore())
	if err != nil { t.Fatal(err) }
	now = now.Add(8 * time.Hour)
	if _, err := service.Snapshot(); err != nil { t.Fatal(err) }
	for i := 0; i < 4; i++ { if _, err := service.UpgradeBusiness("carrot-logistics"); err != nil { t.Fatal(err) } }
	for i := 0; i < 4; i++ { if _, err := service.UpgradeBusiness("scrap-salvage"); err != nil { t.Fatal(err) } }
	snapshot, err := service.Snapshot()
	if err != nil { t.Fatal(err) }
	var logistics BusinessView
	for _, business := range snapshot.Businesses { if business.ID == "carrot-logistics" { logistics = business } }
	if logistics.Level != 5 { t.Fatalf("level=%d want 5", logistics.Level) }
	if logistics.IncomePerSecond != 55 { t.Fatalf("income=%d want 55", logistics.IncomePerSecond) }
	if logistics.Synergy != "+10% synergy" { t.Fatalf("synergy=%q", logistics.Synergy) }
}

func TestSeasonTurnInAwardsCosmeticAndResetsPowerlessEconomy(t *testing.T) {
	now := time.Date(2026, 9, 11, 12, 0, 0, 0, time.UTC)
	service, err := NewService(func() time.Time { return now }, NewMemoryStore())
	if err != nil { t.Fatal(err) }
	now = now.Add(30 * time.Minute)
	if _, err := service.Snapshot(); err != nil { t.Fatal(err) }
	turned, err := service.TurnInSeason()
	if err != nil { t.Fatal(err) }
	if turned.Season.TurnIns != 1 || turned.Season.Earnings != 0 { t.Fatalf("season=%+v", turned.Season) }
	if turned.Wallet.BunnyBucks != 750 { t.Fatalf("wallet=%d", turned.Wallet.BunnyBucks) }
	if !contains(turned.Player.Cosmetics, "Season Zero Pennant") { t.Fatalf("cosmetics=%v", turned.Player.Cosmetics) }
	if turned.RankedPowerAffected { t.Fatal("idle economy must never affect ranked combat power") }
}

func TestProfileValidationAndPersistence(t *testing.T) {
	now := time.Date(2026, 9, 11, 12, 0, 0, 0, time.UTC)
	path := filepath.Join(t.TempDir(), "game-state.json")
	store := NewFileStore(path)
	service, err := NewService(func() time.Time { return now }, store)
	if err != nil { t.Fatal(err) }
	profile := PlayerProfile{Name: "Boomtail", Callsign: "Fuse", Fur: "charcoal", Ears: "battle-worn", Uniform: "night-black"}
	if _, err := service.UpdateProfile(profile); err != nil { t.Fatal(err) }
	reloaded, err := NewService(func() time.Time { return now }, store)
	if err != nil { t.Fatal(err) }
	snapshot, err := reloaded.Snapshot()
	if err != nil { t.Fatal(err) }
	if snapshot.Player.Name != "Boomtail" || snapshot.Player.Callsign != "Fuse" { t.Fatalf("player=%+v", snapshot.Player) }
	if !contains(snapshot.Player.Cosmetics, "Recruit Patch") { t.Fatal("profile edit must preserve cosmetics") }
}

func TestOnboardingProgressPersists(t *testing.T) {
	now := time.Date(2026, 9, 11, 12, 0, 0, 0, time.UTC)
	store := NewMemoryStore()
	service, err := NewService(func() time.Time { return now }, store)
	if err != nil { t.Fatal(err) }
	for i := 0; i < 8; i++ { if _, err := service.AdvanceOnboarding(); err != nil { t.Fatal(err) } }
	snapshot, err := service.Snapshot()
	if err != nil { t.Fatal(err) }
	if snapshot.OnboardingStep != 6 { t.Fatalf("onboarding=%d want 6", snapshot.OnboardingStep) }
}
