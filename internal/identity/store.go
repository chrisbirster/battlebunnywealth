package identity

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

var ErrNotFound = errors.New("identity state not found")

type Store interface {
	Load() (State, error)
	Save(State) error
}

type FileStore struct{ Path string }

func NewFileStore(path string) *FileStore { return &FileStore{Path: path} }

func (s *FileStore) Load() (State, error) {
	data, err := os.ReadFile(s.Path)
	if errors.Is(err, os.ErrNotExist) {
		return State{}, ErrNotFound
	}
	if err != nil {
		return State{}, fmt.Errorf("read identity state: %w", err)
	}
	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return State{}, fmt.Errorf("decode identity state: %w", err)
	}
	return state, nil
}

func (s *FileStore) Save(state State) error {
	if err := os.MkdirAll(filepath.Dir(s.Path), 0o755); err != nil {
		return fmt.Errorf("create identity state directory: %w", err)
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("encode identity state: %w", err)
	}
	tmp := s.Path + ".tmp"
	if err := os.WriteFile(tmp, append(data, '\n'), 0o600); err != nil {
		return fmt.Errorf("write identity state: %w", err)
	}
	if err := os.Rename(tmp, s.Path); err != nil {
		return fmt.Errorf("replace identity state: %w", err)
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
	state.Accounts = append([]Account(nil), state.Accounts...)
	for i := range state.Accounts {
		state.Accounts[i].Credentials = append([]PasskeyCredential(nil), state.Accounts[i].Credentials...)
		for j := range state.Accounts[i].Credentials {
			state.Accounts[i].Credentials[j].Transports = append([]string(nil), state.Accounts[i].Credentials[j].Transports...)
		}
		state.Accounts[i].Devices = append([]Device(nil), state.Accounts[i].Devices...)
		if state.Accounts[i].ATProto != nil {
			binding := *state.Accounts[i].ATProto
			state.Accounts[i].ATProto = &binding
		}
	}
	state.Sessions = append([]Session(nil), state.Sessions...)
	return state
}
