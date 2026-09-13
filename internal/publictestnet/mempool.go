package publictestnet

import (
	"sort"
	"sync"

	"github.com/chrisbirster/battlebunnywealth/internal/testnet"
)

type Mempool struct{mu sync.Mutex;networkID string;max int;entries map[string]SignedTransaction;bySender map[string]string}
func NewMempool(networkID string,maxEntries int)*Mempool{if maxEntries<=0{maxEntries=4096};return &Mempool{networkID:networkID,max:maxEntries,entries:map[string]SignedTransaction{},bySender:map[string]string{}}}
func(m *Mempool)Admit(tx SignedTransaction,currentHeight,nextNonce uint64)error{m.mu.Lock();defer m.mu.Unlock();if _,ok:=m.entries[tx.ID];ok{return ErrDuplicateTransaction};if _,ok:=m.bySender[tx.From];ok{return ErrInvalidNonce};if len(m.entries)>=m.max{return ErrMempoolFull};if err:=ValidateTransaction(tx,m.networkID,currentHeight,nextNonce);err!=nil{return err};m.entries[tx.ID]=tx;m.bySender[tx.From]=tx.ID;return nil}
func(m *Mempool)Remove(id string){m.mu.Lock();defer m.mu.Unlock();if tx,ok:=m.entries[id];ok{delete(m.bySender,tx.From)};delete(m.entries,id)}
func(m *Mempool)Count()int{m.mu.Lock();defer m.mu.Unlock();return len(m.entries)}
func(m *Mempool)Transactions()[]SignedTransaction{m.mu.Lock();defer m.mu.Unlock();out:=make([]SignedTransaction,0,len(m.entries));for _,tx:=range m.entries{out=append(out,tx)};sort.Slice(out,func(i,j int)bool{return out[i].ID<out[j].ID});return out}
func(m *Mempool)Operations(max int)[]testnet.Operation{txs:=m.Transactions();if max<=0||max>len(txs){max=len(txs)};ops:=make([]testnet.Operation,0,max);for _,tx:=range txs[:max]{op,err:=TransactionOperation(tx);if err==nil{ops=append(ops,op)}};return ops}
func(m *Mempool)RemoveOperations(ops []testnet.Operation){for _,op:=range ops{if op.Type!=TransactionOperationType{continue};tx,err:=ParseTransactionOperation(op);if err==nil{m.Remove(tx.ID)}}}
