package game

import (
	"path/filepath"
	"testing"
	"time"
)

func TestRegistryIsolatesAccountState(t *testing.T) {
	now := time.Date(2026, 9, 11, 12, 0, 0, 0, time.UTC)
	registry := NewRegistry(func() time.Time { return now }, filepath.Join(t.TempDir(), "players"), "")
	first, err := registry.ForAccount("abcdefghijklmnopqrstuvwx")
	if err != nil { t.Fatal(err) }
	second, err := registry.ForAccount("zyxwvutsrqponmlkjihgfedc")
	if err != nil { t.Fatal(err) }
	if _, err := first.UpdateProfile(PlayerProfile{Name:"Boomtail",Callsign:"Fuse",Fur:"charcoal",Ears:"battle-worn",Uniform:"night-black"}); err != nil { t.Fatal(err) }
	one, _ := first.Snapshot(); two, _ := second.Snapshot()
	if one.Player.Name != "Boomtail" { t.Fatalf("first player=%q", one.Player.Name) }
	if two.Player.Name == "Boomtail" { t.Fatal("account game states leaked into each other") }
}
