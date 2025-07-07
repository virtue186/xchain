package core

import (
	"crypto/sha256"
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

func (TxHasher) Hash(transaction *Transaction) types.Hash {
	return sha256.Sum256(transaction.Data)
}
