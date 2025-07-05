package core

import (
	"fmt"
	"github.com/virtue186/xchain/crypto"
	"io"
)

type Transaction struct {
	Data      []byte
	PublicKey crypto.PublicKey
	Signature *crypto.Signature
}

func (tx *Transaction) EncodeBinary(r io.Writer) error {
	return nil
}

func (tx *Transaction) DecodeBinary(r io.Reader) error {
	return nil
}

func (tx *Transaction) Sign(privateKey crypto.PrivateKey) error {
	sign, err := privateKey.Sign(tx.Data)
	if err != nil {
		return err
	}
	tx.PublicKey = privateKey.PublicKey()
	tx.Signature = sign
	return nil
}

func (tx *Transaction) Verify() error {
	if tx.Signature == nil {
		return fmt.Errorf("transaction signature is nil")
	}

	if !tx.Signature.Verify(tx.PublicKey, tx.Data) {
		return fmt.Errorf("transaction signature is invalid")
	}
	return nil
}
