package publictestnet

import("sync";"github.com/chrisbirster/battlebunnywealth/internal/carrot")
type PublicLedger struct{mu sync.Mutex;ledger *carrot.Ledger;nonces map[string]uint64}
func NewPublicLedger(policy carrot.Policy)(*PublicLedger,error){l,err:=carrot.NewLedger(policy);if err!=nil{return nil,err};return &PublicLedger{ledger:l,nonces:map[string]uint64{}},nil}
func(l *PublicLedger)FundAtHeight(height uint64,recipients []string)(map[string]int64,error){l.mu.Lock();defer l.mu.Unlock();return l.ledger.ReleaseParticipation(height,recipients)}
func(l *PublicLedger)Apply(tx SignedTransaction,networkID string,currentHeight uint64,feeRecipients []string)error{l.mu.Lock();defer l.mu.Unlock();expected:=l.nonces[tx.From];if err:=ValidateTransaction(tx,networkID,currentHeight,expected);err!=nil{return err};if err:=l.ledger.Transfer(tx.From,tx.To,tx.AmountAtoms,tx.FeeAtoms,feeRecipients);err!=nil{return err};l.nonces[tx.From]=expected+1;return l.ledger.ValidateConservation()}
func(l *PublicLedger)Balance(a string)int64{l.mu.Lock();defer l.mu.Unlock();return l.ledger.Balance(a)}
func(l *PublicLedger)Nonce(a string)uint64{l.mu.Lock();defer l.mu.Unlock();return l.nonces[a]}
func(l *PublicLedger)SupplyReport()carrot.SupplyReport{l.mu.Lock();defer l.mu.Unlock();return l.ledger.SupplyReport()}
func(l *PublicLedger)ValidateConservation()error{l.mu.Lock();defer l.mu.Unlock();return l.ledger.ValidateConservation()}
