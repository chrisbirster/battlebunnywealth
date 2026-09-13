package publictestnet

import(
"encoding/json";"errors";"github.com/chrisbirster/battlebunnywealth/internal/testnet")
const TransactionOperationType="test-carrot-transfer"
func TransactionOperation(tx SignedTransaction)(testnet.Operation,error){if tx.ID==""{return testnet.Operation{},ErrInvalidTransaction};raw,err:=json.Marshal(tx);if err!=nil{return testnet.Operation{},err};return testnet.Operation{Type:TransactionOperationType,Key:tx.ID,Value:string(raw)},nil}
func ParseTransactionOperation(op testnet.Operation)(SignedTransaction,error){if op.Type!=TransactionOperationType||op.Key==""||op.Value==""{return SignedTransaction{},errors.New("not a TEST-CARROT transfer operation")};var tx SignedTransaction;if err:=json.Unmarshal([]byte(op.Value),&tx);err!=nil{return SignedTransaction{},err};if tx.ID!=op.Key{return SignedTransaction{},ErrInvalidTransaction};return tx,nil}
