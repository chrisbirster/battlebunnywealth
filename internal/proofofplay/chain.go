package proofofplay

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrInvalidHeight       = errors.New("invalid block height")
	ErrInvalidPreviousHash = errors.New("invalid previous hash")
	ErrInvalidBlockHash    = errors.New("invalid block hash")
	ErrEpochRegression     = errors.New("block epoch regressed")
)

type Chain struct {
	mu     sync.RWMutex
	blocks []Block
}

func NewChain(genesisAt time.Time) *Chain {
	return &Chain{blocks: []Block{NewGenesis(genesisAt)}}
}

func (c *Chain) Head() Block {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return cloneBlock(c.blocks[len(c.blocks)-1])
}

func (c *Chain) Height() uint64 {
	return c.Head().Header.Height
}

func (c *Chain) Append(block Block) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	head := c.blocks[len(c.blocks)-1]
	if block.Header.Height != head.Header.Height+1 {
		return fmt.Errorf("%w: got %d want %d", ErrInvalidHeight, block.Header.Height, head.Header.Height+1)
	}
	if block.Header.PreviousHash != head.Hash {
		return ErrInvalidPreviousHash
	}
	if block.Header.Epoch < head.Header.Epoch {
		return ErrEpochRegression
	}
	if block.Hash != block.CalculateHash() {
		return ErrInvalidBlockHash
	}
	c.blocks = append(c.blocks, cloneBlock(block))
	return nil
}

func cloneBlock(block Block) Block {
	block.Transactions = append([]Transaction(nil), block.Transactions...)
	block.Proofs = append([]ParticipationProof(nil), block.Proofs...)
	return block
}
