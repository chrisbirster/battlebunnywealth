package testnet

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type Store struct{ Dir string }

func OpenStore(dir string, genesis Genesis) (*Store, error) {
	if dir == "" {
		return nil, errors.New("data directory required")
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	s := &Store{Dir: dir}
	path := filepath.Join(dir, "genesis.json")
	if raw, err := os.ReadFile(path); err == nil {
		var existing Genesis
		if json.Unmarshal(raw, &existing) != nil || existing.Hash != genesis.Hash {
			return nil, errors.New("data directory genesis mismatch")
		}
	} else if errors.Is(err, os.ErrNotExist) {
		raw, _ := json.MarshalIndent(genesis, "", "  ")
		if err := atomicWrite(path, append(raw, '\n'), 0o600); err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}
	return s, nil
}
func (s *Store) Append(block FinalizedBlock) error {
	path := filepath.Join(s.Dir, "blocks.ndjson")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	raw, err := json.Marshal(block)
	if err != nil {
		return err
	}
	if _, err = f.Write(append(raw, '\n')); err != nil {
		return err
	}
	return f.Sync()
}
func (s *Store) Load() ([]FinalizedBlock, error) {
	path := filepath.Join(s.Dir, "blocks.ndjson")
	f, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var out []FinalizedBlock
	scan := bufio.NewScanner(f)
	buf := make([]byte, 64*1024)
	scan.Buffer(buf, 4<<20)
	line := 0
	for scan.Scan() {
		line++
		var block FinalizedBlock
		if err := json.Unmarshal(scan.Bytes(), &block); err != nil {
			return nil, fmt.Errorf("decode block line %d: %w", line, err)
		}
		out = append(out, block)
	}
	return out, scan.Err()
}
func (s *Store) Restore(engine *Engine) error {
	blocks, err := s.Load()
	if err != nil {
		return err
	}
	for _, block := range blocks {
		if err := engine.ImportFinalized(block); err != nil {
			return fmt.Errorf("verify stored block %d: %w", block.Block.Height, err)
		}
	}
	return nil
}
func atomicWrite(path string, data []byte, mode os.FileMode) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, mode); err != nil {
		return err
	}
	if err := os.Chmod(tmp, mode); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
