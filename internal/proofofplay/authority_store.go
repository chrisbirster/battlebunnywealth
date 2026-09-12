package proofofplay

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

var ErrAuthorityStateNotFound = errors.New("proof-of-play authority state not found")

type AuthorityStore interface {
	Load() (AuthorityStoreState, error)
	Save(AuthorityStoreState) error
}

type AuthorityFileStore struct{ Path string }

func NewAuthorityFileStore(path string) *AuthorityFileStore { return &AuthorityFileStore{Path: path} }

func (s *AuthorityFileStore) Load() (AuthorityStoreState, error) {
	data, err := os.ReadFile(s.Path)
	if errors.Is(err, os.ErrNotExist) {
		return AuthorityStoreState{}, ErrAuthorityStateNotFound
	}
	if err != nil {
		return AuthorityStoreState{}, fmt.Errorf("read proof-of-play authority state: %w", err)
	}
	var state AuthorityStoreState
	if err := json.Unmarshal(data, &state); err != nil {
		return AuthorityStoreState{}, fmt.Errorf("decode proof-of-play authority state: %w", err)
	}
	return state, nil
}

func (s *AuthorityFileStore) Save(state AuthorityStoreState) error {
	if err := os.MkdirAll(filepath.Dir(s.Path), 0o755); err != nil {
		return fmt.Errorf("create proof-of-play state directory: %w", err)
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("encode proof-of-play authority state: %w", err)
	}
	tmp := s.Path + ".tmp"
	if err := os.WriteFile(tmp, append(data, '\n'), 0o600); err != nil {
		return fmt.Errorf("write proof-of-play authority state: %w", err)
	}
	if err := os.Rename(tmp, s.Path); err != nil {
		return fmt.Errorf("replace proof-of-play authority state: %w", err)
	}
	return nil
}

type MemoryAuthorityStore struct {
	mu    sync.Mutex
	state *AuthorityStoreState
}

func NewMemoryAuthorityStore() *MemoryAuthorityStore { return &MemoryAuthorityStore{} }

func (s *MemoryAuthorityStore) Load() (AuthorityStoreState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.state == nil {
		return AuthorityStoreState{}, ErrAuthorityStateNotFound
	}
	return cloneAuthorityStoreState(*s.state), nil
}

func (s *MemoryAuthorityStore) Save(state AuthorityStoreState) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	copy := cloneAuthorityStoreState(state)
	s.state = &copy
	return nil
}

func cloneAuthorityStoreState(state AuthorityStoreState) AuthorityStoreState {
	state.Missions = append([]NetworkMission(nil), state.Missions...)
	state.Authority = append([]AuthorityRecord(nil), state.Authority...)
	state.Completions = append([]MissionCompletion(nil), state.Completions...)
	return state
}
