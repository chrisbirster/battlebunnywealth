package publictestnet

import (
	"sort"
	"sync"

	"github.com/chrisbirster/battlebunnywealth/internal/testnet"
)

const (
	MaxQueuedTransactionsPerSender = 16
	ReplacementFeeBumpBasisPoints  = int64(1250) // 12.5%, minimum one atom.
)

type Mempool struct {
	mu        sync.Mutex
	networkID string
	max       int
	entries   map[string]SignedTransaction
	bySender  map[string]map[uint64]string
}

func NewMempool(networkID string, maxEntries int) *Mempool {
	if maxEntries <= 0 {
		maxEntries = 4096
	}
	return &Mempool{
		networkID: networkID,
		max:       maxEntries,
		entries:   map[string]SignedTransaction{},
		bySender:  map[string]map[uint64]string{},
	}
}

func (m *Mempool) Admit(tx SignedTransaction, currentHeight, nextNonce uint64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.entries[tx.ID]; ok {
		return ErrDuplicateTransaction
	}
	if err := ValidateTransactionEnvelope(tx, m.networkID, currentHeight); err != nil {
		return err
	}
	if tx.Nonce < nextNonce || tx.Nonce-nextNonce >= MaxQueuedTransactionsPerSender {
		return ErrInvalidNonce
	}

	queue := m.bySender[tx.From]
	if queue == nil {
		queue = map[uint64]string{}
		m.bySender[tx.From] = queue
	}
	if priorID, exists := queue[tx.Nonce]; exists {
		prior := m.entries[priorID]
		minimumBump := prior.FeeAtoms * ReplacementFeeBumpBasisPoints / 10_000
		if prior.FeeAtoms*ReplacementFeeBumpBasisPoints%10_000 != 0 {
			minimumBump++
		}
		if minimumBump < 1 {
			minimumBump = 1
		}
		if tx.FeeAtoms < prior.FeeAtoms+minimumBump {
			return ErrReplacementUnderpriced
		}
		delete(m.entries, priorID)
		m.entries[tx.ID] = tx
		queue[tx.Nonce] = tx.ID
		return nil
	}
	if len(queue) >= MaxQueuedTransactionsPerSender {
		return ErrInvalidNonce
	}
	if len(m.entries) >= m.max {
		return ErrMempoolFull
	}
	m.entries[tx.ID] = tx
	queue[tx.Nonce] = tx.ID
	return nil
}

func (m *Mempool) Remove(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if tx, ok := m.entries[id]; ok {
		if queue := m.bySender[tx.From]; queue != nil {
			delete(queue, tx.Nonce)
			if len(queue) == 0 {
				delete(m.bySender, tx.From)
			}
		}
	}
	delete(m.entries, id)
}

func (m *Mempool) Count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.entries)
}

func (m *Mempool) Transactions() []SignedTransaction {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]SignedTransaction, 0, len(m.entries))
	for _, tx := range m.entries {
		out = append(out, tx)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].From != out[j].From {
			return out[i].From < out[j].From
		}
		if out[i].Nonce != out[j].Nonce {
			return out[i].Nonce < out[j].Nonce
		}
		return out[i].ID < out[j].ID
	})
	return out
}

// Operations preserves the v0.12 API for callers that do not have a ledger
// nonce view. Consensus-facing callers should use OperationsFrom.
func (m *Mempool) Operations(max int) []testnet.Operation {
	txs := m.Transactions()
	if max <= 0 || max > len(txs) {
		max = len(txs)
	}
	ops := make([]testnet.Operation, 0, max)
	for _, tx := range txs[:max] {
		op, err := TransactionOperation(tx)
		if err == nil {
			ops = append(ops, op)
		}
	}
	return ops
}

// OperationsFrom returns only contiguous executable nonce chains, ordered by
// sender then nonce. A queued nonce gap therefore cannot make a proposal
// invalid or block another sender's ready transaction.
func (m *Mempool) OperationsFrom(max int, nextNonce func(string) uint64) []testnet.Operation {
	m.mu.Lock()
	defer m.mu.Unlock()
	if max <= 0 {
		max = len(m.entries)
	}
	senders := make([]string, 0, len(m.bySender))
	for sender := range m.bySender {
		senders = append(senders, sender)
	}
	sort.Strings(senders)
	ops := make([]testnet.Operation, 0, max)
	for _, sender := range senders {
		expected := uint64(0)
		if nextNonce != nil {
			expected = nextNonce(sender)
		}
		queue := m.bySender[sender]
		for len(ops) < max {
			id, ok := queue[expected]
			if !ok {
				break
			}
			tx := m.entries[id]
			op, err := TransactionOperation(tx)
			if err != nil {
				break
			}
			ops = append(ops, op)
			expected++
		}
		if len(ops) >= max {
			break
		}
	}
	return ops
}

func (m *Mempool) RemoveOperations(ops []testnet.Operation) {
	for _, op := range ops {
		if op.Type != TransactionOperationType {
			continue
		}
		tx, err := ParseTransactionOperation(op)
		if err == nil {
			m.Remove(tx.ID)
		}
	}
}
