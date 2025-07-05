package core

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/virtue186/xchain/crypto"
	"github.com/virtue186/xchain/types"
	"io"
)

type Header struct {
	Version       uint32
	PrevBlockHash types.Hash
	DataHash      types.Hash
	Timestamp     int64
	Height        uint32
	Nonce         uint64
}
type Block struct {
	Header
	Transactions []Transaction
	Validator    crypto.PublicKey
	Signature    *crypto.Signature

	// cached version of the header hash
	hash types.Hash
}

func (b *Block) Hash(hasher Hasher[*Block]) types.Hash {

	if b.hash.IsZero() {
		b.hash = hasher.Hash(b)
	}
	return b.hash
}

func (b *Block) Encode(w io.Writer, enc Encoder[*Block]) error {
	return enc.Encode(w, b)
}

func (b *Block) Decode(r io.Reader, enc Decoder[*Block]) error {
	return enc.Decode(r, b)
}

func NewBlock(header *Header, transactions []Transaction) *Block {
	return &Block{
		Header:       *header,
		Transactions: transactions,
	}
}

func (b *Block) Sign(privateKey crypto.PrivateKey) error {
	sign, err := privateKey.Sign(b.HeaderData())
	if err != nil {
		return err
	}
	b.Validator = privateKey.PublicKey()
	b.Signature = sign
	return nil
}

func (b *Block) Verify() error {
	if b.Signature == nil {
		return fmt.Errorf("signature is nil")
	}
	if !b.Signature.Verify(b.Validator, b.HeaderData()) {
		return fmt.Errorf("signature is invalid")
	}
	return nil
}

func (b *Block) HeaderData() []byte {
	buf := &bytes.Buffer{}
	err := gob.NewEncoder(buf).Encode(b.Header)
	if err != nil {
		return nil
	}
	return buf.Bytes()
}
