package core

import (
	"fmt"
	"github.com/virtue186/xchain/types"
	"testing"
	"time"
)

func RandomBlock(height uint32) *Block {

	h := &Header{
		Version:       1,
		PrevBlockHash: types.RandomHash(),
		DataHash:      types.RandomHash(),
		Timestamp:     time.Now().UnixNano(),
		Height:        height,
		Nonce:         8184848,
	}
	t := Transaction{
		Data: []byte("hello world"),
	}
	return NewBlock(h, []Transaction{t})
}

func TestBlock_Hash(t *testing.T) {
	block := RandomBlock(0)
	fmt.Println(block.Hash(BlockHasher{}))
}
