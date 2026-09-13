package publictestnet

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
)

const TransactionVersion = 1
const MaxTransactionTTL uint64 = 1008

var (
	ErrWrongTransactionNetwork = errors.New("wrong TEST-CARROT transaction network")
	ErrInvalidTransaction      = errors.New("invalid TEST-CARROT transaction")
	ErrInvalidTransactionSig   = errors.New("invalid TEST-CARROT transaction signature")
	ErrTransactionExpired      = errors.New("TEST-CARROT transaction expired")
	ErrInvalidNonce            = errors.New("invalid TEST-CARROT nonce")
	ErrDuplicateTransaction    = errors.New("duplicate TEST-CARROT transaction")
	ErrMempoolFull             = errors.New("TEST-CARROT mempool full")
	ErrReplacementUnderpriced  = errors.New("replacement TEST-CARROT transaction fee bump too small")
)

type SignedTransaction struct {
	Version     int    `json:"version"`
	NetworkID   string `json:"networkId"`
	From        string `json:"from"`
	To          string `json:"to"`
	AmountAtoms int64  `json:"amountAtoms"`
	FeeAtoms    int64  `json:"feeAtoms"`
	Nonce       uint64 `json:"nonce"`
	ValidUntil  uint64 `json:"validUntil"`
	PublicKey   string `json:"publicKey"`
	Signature   string `json:"signature"`
	ID          string `json:"id"`
}

func NewSignedTransaction(networkID, to string, amountAtoms, feeAtoms int64, nonce, validUntil uint64, w Wallet) (SignedTransaction, error) {
	if len(w.Private) != ed25519.PrivateKeySize {
		return SignedTransaction{}, errors.New("wallet private key required")
	}
	tx := SignedTransaction{
		Version:     TransactionVersion,
		NetworkID:   networkID,
		From:        w.Address(),
		To:          to,
		AmountAtoms: amountAtoms,
		FeeAtoms:    feeAtoms,
		Nonce:       nonce,
		ValidUntil:  validUntil,
		PublicKey:   w.PublicText(),
	}
	tx.Signature = base64.RawURLEncoding.EncodeToString(ed25519.Sign(w.Private, []byte(transactionSigningMessage(tx))))
	tx.ID = transactionID(tx)
	return tx, nil
}

// ValidateTransactionEnvelope verifies every transaction invariant except the
// state-dependent exact nonce. Mempool queueing uses this to admit a bounded
// sequence of future nonces while consensus still enforces exact sequential
// execution with ValidateTransaction.
func ValidateTransactionEnvelope(tx SignedTransaction, networkID string, currentHeight uint64) error {
	if tx.Version != TransactionVersion || tx.NetworkID == "" || tx.NetworkID != networkID {
		return ErrWrongTransactionNetwork
	}
	if tx.AmountAtoms <= 0 || tx.FeeAtoms < 0 || tx.From == tx.To || !ValidAddress(tx.To) {
		return ErrInvalidTransaction
	}
	from, err := AddressFromPublicText(tx.PublicKey)
	if err != nil || from != tx.From {
		return ErrInvalidTransaction
	}
	if tx.ValidUntil <= currentHeight || tx.ValidUntil-currentHeight > MaxTransactionTTL {
		return ErrTransactionExpired
	}
	pub, err := base64.RawURLEncoding.DecodeString(tx.PublicKey)
	if err != nil || len(pub) != ed25519.PublicKeySize {
		return ErrInvalidTransactionSig
	}
	sig, err := base64.RawURLEncoding.DecodeString(tx.Signature)
	if err != nil || !ed25519.Verify(ed25519.PublicKey(pub), []byte(transactionSigningMessage(tx)), sig) {
		return ErrInvalidTransactionSig
	}
	if transactionID(tx) != tx.ID {
		return ErrInvalidTransactionSig
	}
	return nil
}

func ValidateTransaction(tx SignedTransaction, networkID string, currentHeight, expectedNonce uint64) error {
	if err := ValidateTransactionEnvelope(tx, networkID, currentHeight); err != nil {
		return err
	}
	if tx.Nonce != expectedNonce {
		return ErrInvalidNonce
	}
	return nil
}

func transactionSigningMessage(tx SignedTransaction) string {
	return fmt.Sprintf("bbw-test-carrot-tx/v1\nnetwork=%s\nfrom=%s\nto=%s\namount=%d\nfee=%d\nnonce=%d\nvalidUntil=%d\npublicKey=%s", tx.NetworkID, tx.From, tx.To, tx.AmountAtoms, tx.FeeAtoms, tx.Nonce, tx.ValidUntil, tx.PublicKey)
}

func transactionID(tx SignedTransaction) string {
	sum := sha256.Sum256([]byte(transactionSigningMessage(tx) + "\nsignature=" + tx.Signature))
	return hex.EncodeToString(sum[:])
}
