package core

import (
	"fmt"
	"github.com/virtue186/xchain/crypto"
	"github.com/virtue186/xchain/types"
)

type Transaction struct {
	Data      []byte
	From      crypto.PublicKey
	Signature *crypto.Signature

	hash      types.Hash
	firstSeen int64
}

func NewTransaction(data []byte) *Transaction {
	return &Transaction{
		Data: data,
	}
}

func (tx *Transaction) Hash(hasher Hasher[*Transaction]) types.Hash {
	if tx.hash.IsZero() {
		tx.hash = hasher.Hash(tx)
	}
	return tx.hash
}

func (tx *Transaction) Sign(privateKey crypto.PrivateKey) error {
	sign, err := privateKey.Sign(tx.Data)
	if err != nil {
		return err
	}
	tx.From = privateKey.PublicKey()
	tx.Signature = sign
	return nil
}

func (tx *Transaction) Verify() error {
	if tx.Signature == nil {
		return fmt.Errorf("transaction signature is nil")
	}

	if !tx.Signature.Verify(tx.From, tx.Data) {
		return fmt.Errorf("transaction signature is invalid")
	}
	return nil
}

func (tx *Transaction) SetFirstSeen(firstSeen int64) {
	tx.firstSeen = firstSeen
}

func (tx *Transaction) GetFirstSeen() int64 {
	return tx.firstSeen
}
