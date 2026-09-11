package proofofplay

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

type BlockHeader struct {
	Version      uint16    `json:"version"`
	Height       uint64    `json:"height"`
	PreviousHash string    `json:"previousHash"`
	Epoch        uint64    `json:"epoch"`
	Timestamp    time.Time `json:"timestamp"`
	Proposer     string    `json:"proposer"`
}

type Block struct {
	Header       BlockHeader          `json:"header"`
	Transactions []Transaction        `json:"transactions"`
	Proofs       []ParticipationProof `json:"proofs"`
	Hash         string               `json:"hash"`
}

func NewGenesis(at time.Time) Block {
	block := Block{
		Header: BlockHeader{
			Version:      1,
			Height:       0,
			PreviousHash: stringsOfZero(64),
			Epoch:        0,
			Timestamp:    at.UTC(),
			Proposer:     "genesis",
		},
		Transactions: []Transaction{},
		Proofs:       []ParticipationProof{},
	}
	block.Hash = block.CalculateHash()
	return block
}

func NewBlock(previous Block, epoch uint64, proposer string, txs []Transaction, proofs []ParticipationProof, at time.Time) Block {
	block := Block{
		Header: BlockHeader{
			Version:      1,
			Height:       previous.Header.Height + 1,
			PreviousHash: previous.Hash,
			Epoch:        epoch,
			Timestamp:    at.UTC(),
			Proposer:     proposer,
		},
		Transactions: append([]Transaction(nil), txs...),
		Proofs:       append([]ParticipationProof(nil), proofs...),
	}
	block.Hash = block.CalculateHash()
	return block
}

func (b Block) CalculateHash() string {
	payload := struct {
		Header       BlockHeader
		Transactions []Transaction
		Proofs       []ParticipationProof
	}{b.Header, b.Transactions, b.Proofs}

	raw, err := json.Marshal(payload)
	if err != nil {
		panic(fmt.Sprintf("marshal deterministic block payload: %v", err))
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func stringsOfZero(n int) string {
	buf := make([]byte, n)
	for i := range buf {
		buf[i] = '0'
	}
	return string(buf)
}
