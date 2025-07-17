package core

import (
	"crypto/sha256"
	"encoding/json"
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

	b, err := json.Marshal(tx)
	if err != nil {
		panic(err)
	}
	return sha256.Sum256(b)
}
