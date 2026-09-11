package game

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

var ErrNotFound = errors.New("game state not found")

type Store interface {
	Load() (State, error)
	Save(State) error
}

type FileStore struct {
	Path string
}

func NewFileStore(path string) *FileStore { return &FileStore{Path: path} }

func (s *FileStore) Load() (State, error) {
	data, err := os.ReadFile(s.Path)
	if errors.Is(err, os.ErrNotExist) {
		return State{}, ErrNotFound
	}
	if err != nil {
		return State{}, fmt.Errorf("read game state: %w", err)
	}
	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return State{}, fmt.Errorf("decode game state: %w", err)
	}
	return state, nil
}

func (s *FileStore) Save(state State) error {
	if err := os.MkdirAll(filepath.Dir(s.Path), 0o755); err != nil {
		return fmt.Errorf("create game state directory: %w", err)
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("encode game state: %w", err)
	}
	tmp := s.Path + ".tmp"
	if err := os.WriteFile(tmp, append(data, '\n'), 0o600); err != nil {
		return fmt.Errorf("write game state: %w", err)
	}
	if err := os.Rename(tmp, s.Path); err != nil {
		return fmt.Errorf("replace game state: %w", err)
	}
	return nil
}

type MemoryStore struct {
	mu    sync.Mutex
	state *State
}

func NewMemoryStore() *MemoryStore { return &MemoryStore{} }

func (s *MemoryStore) Load() (State, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.state == nil {
		return State{}, ErrNotFound
	}
	return cloneState(*s.state), nil
}

func (s *MemoryStore) Save(state State) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	copy := cloneState(state)
	s.state = &copy
	return nil
}

func cloneState(state State) State {
	state.Player.Cosmetics = append([]string(nil), state.Player.Cosmetics...)
	state.Businesses = append([]BusinessState(nil), state.Businesses...)
	return state
}
