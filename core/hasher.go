package core

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"github.com/virtue186/xchain/types"
)

type Hasher[T any] interface {
	Hash(T) types.Hash
}

type BlockHasher struct {
}

func (BlockHasher) Hash(h *Header) types.Hash {

	sum256 := sha256.Sum256(h.Bytes())
	return sum256
}

type TxHasher struct {
}

// Hash 计算交易的哈希值
func (TxHasher) Hash(tx *Transaction) types.Hash {
	dataToHash := &struct {
		Data  []byte
		To    types.Address
		Value uint64
		Nonce uint64
	}{
		To:    tx.To,
		Value: tx.Value,
		Nonce: tx.Nonce,
	}

	if len(tx.Data) == 0 {
		dataToHash.Data = []byte{}
	} else {
		dataToHash.Data = tx.Data
	}

	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(dataToHash); err != nil {
		panic(err)
	}
	return sha256.Sum256(buf.Bytes())
}
