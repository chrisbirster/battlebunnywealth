package publictestnet

import (
	"encoding/json"
	"errors"
	"sort"

	"github.com/chrisbirster/battlebunnywealth/internal/testnet"
)

const (
	TransactionOperationType = "test-carrot-transfer"
	SettlementOperationType  = "test-carrot-finality-settlement"
)

type FinalitySettlement struct {
	RewardHeight uint64         `json:"rewardHeight"`
	BlockHash    string         `json:"blockHash"`
	Round        uint32         `json:"round"`
	Votes        []testnet.Vote `json:"votes"`
}

func TransactionOperation(tx SignedTransaction) (testnet.Operation, error) {
	if tx.ID == "" { return testnet.Operation{}, ErrInvalidTransaction }
	raw, err := json.Marshal(tx); if err != nil { return testnet.Operation{}, err }
	return testnet.Operation{Type:TransactionOperationType,Key:tx.ID,Value:string(raw)},nil
}
func ParseTransactionOperation(op testnet.Operation) (SignedTransaction,error) {
	if op.Type!=TransactionOperationType||op.Key==""||op.Value==""{return SignedTransaction{},errors.New("not a TEST-CARROT transfer operation")}
	var tx SignedTransaction;if err:=json.Unmarshal([]byte(op.Value),&tx);err!=nil{return SignedTransaction{},err};if tx.ID!=op.Key{return SignedTransaction{},ErrInvalidTransaction};return tx,nil
}
func SettlementOperation(finalized testnet.FinalizedBlock) (testnet.Operation,error) {
	votes:=append([]testnet.Vote(nil),finalized.Certificate.Votes...);sort.Slice(votes,func(i,j int)bool{return votes[i].ValidatorID<votes[j].ValidatorID})
	s:=FinalitySettlement{RewardHeight:finalized.Block.Height,BlockHash:finalized.Block.Hash,Round:finalized.Block.Round,Votes:votes};raw,err:=json.Marshal(s);if err!=nil{return testnet.Operation{},err};return testnet.Operation{Type:SettlementOperationType,Key:finalized.Block.Hash,Value:string(raw)},nil
}
func ParseSettlementOperation(op testnet.Operation)(FinalitySettlement,error){if op.Type!=SettlementOperationType||op.Key==""||op.Value==""{return FinalitySettlement{},errors.New("not a TEST-CARROT finality settlement")};var s FinalitySettlement;if err:=json.Unmarshal([]byte(op.Value),&s);err!=nil{return FinalitySettlement{},err};if s.BlockHash!=op.Key{return FinalitySettlement{},errors.New("settlement block hash mismatch")};return s,nil}
