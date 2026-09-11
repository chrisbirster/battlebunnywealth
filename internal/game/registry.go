package game

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

var accountIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{20,80}$`)

type Registry struct {
	mu         sync.Mutex
	now        func() time.Time
	dir        string
	legacyPath string
	services   map[string]*Service
}

func NewRegistry(now func() time.Time, dir, legacyPath string) *Registry {
	if now == nil { now = time.Now }
	return &Registry{now: now, dir: dir, legacyPath: legacyPath, services: map[string]*Service{}}
}

func (r *Registry) ForAccount(accountID string) (*Service, error) {
	if !accountIDPattern.MatchString(accountID) { return nil, fmt.Errorf("invalid account id") }
	r.mu.Lock(); defer r.mu.Unlock()
	if service := r.services[accountID]; service != nil { return service, nil }
	if err := os.MkdirAll(r.dir, 0o755); err != nil { return nil, fmt.Errorf("create player state directory: %w", err) }
	path := filepath.Join(r.dir, accountID+".json")
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) && r.legacyPath != "" {
		if _, legacyErr := os.Stat(r.legacyPath); legacyErr == nil {
			if err := os.Rename(r.legacyPath, path); err != nil { return nil, fmt.Errorf("claim legacy game state: %w", err) }
		}
	}
	service, err := NewService(r.now, NewFileStore(path)); if err != nil { return nil, err }
	r.services[accountID] = service
	return service, nil
}

func (r *Registry) Standings() ([]Standing, error) {
	r.mu.Lock(); if err := os.MkdirAll(r.dir, 0o755); err != nil { r.mu.Unlock(); return nil, err }; entries, err := os.ReadDir(r.dir); r.mu.Unlock(); if err != nil { return nil, err }
	var all []Standing
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" { continue }
		accountID := strings.TrimSuffix(entry.Name(), ".json")
		service, err := r.ForAccount(accountID); if err != nil { return nil, err }
		rows, err := service.Standings(); if err != nil { return nil, err }
		if len(rows) > 0 { all = append(all, rows[0]) }
	}
	sort.SliceStable(all, func(i, j int) bool { return all[i].BunnyBucks > all[j].BunnyBucks })
	for i := range all { all[i].Rank = i + 1 }
	return all, nil
}
