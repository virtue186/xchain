package core

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"github.com/virtue186/xchain/crypto"
	"github.com/virtue186/xchain/types"
	"time"
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
	*Header
	Transactions []*Transaction
	Validator    crypto.PublicKey
	Signature    *crypto.Signature

	// cached version of the header hash
	hash types.Hash
}

func (h *Header) Bytes() []byte {
	buf := &bytes.Buffer{}
	err := gob.NewEncoder(buf).Encode(h)
	if err != nil {
		return nil
	}
	return buf.Bytes()

}

func (b *Block) Hash(hasher Hasher[*Header]) types.Hash {

	if b.hash.IsZero() {
		b.hash = hasher.Hash(b.Header)
	}
	return b.hash
}

func NewBlock(header *Header, transactions []*Transaction) (*Block, error) {
	return &Block{
		Header:       header,
		Transactions: transactions,
	}, nil
}

func (b *Block) Sign(privateKey crypto.PrivateKey) error {
	sign, err := privateKey.Sign(b.Header.Bytes())
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
	if !b.Signature.Verify(b.Validator, b.Header.Bytes()) {
		return fmt.Errorf("block {%s} signature is invalid", b.Hash(BlockHasher{}))
	}

	for _, tx := range b.Transactions {
		if err := tx.Verify(); err != nil {
			return err
		}
	}

	datahash, err := CalculateDataHash(b.Transactions)
	if err != nil {
		return err
	}
	if datahash != b.DataHash {
		return fmt.Errorf("data hash is invalid")
	}
	return nil
}

func NewBlockFromPreHeader(h *Header, txx []*Transaction) (*Block, error) {
	datahash, err := CalculateDataHash(txx)
	if err != nil {
		return nil, err
	}
	header := &Header{
		Version:       h.Version,
		DataHash:      datahash,
		PrevBlockHash: BlockHasher{}.Hash(h),
		Timestamp:     time.Now().UnixNano(),
		Height:        h.Height + 1,
	}
	return NewBlock(header, txx)
}

func CalculateDataHash(txx []*Transaction) (hash types.Hash, err error) {
	buf := &bytes.Buffer{}

	encoder := GOBEncoder[any]{}
	for _, tx := range txx {
		if err = encoder.Encode(buf, tx); err != nil {
			return
		}
	}

	hash = sha256.Sum256(buf.Bytes())

	return
}
