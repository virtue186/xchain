package core

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/virtue186/xchain/crypto"
	"github.com/virtue186/xchain/types"
	"testing"
	"time"
)

func RandomBlock(height uint32, prevblockhash types.Hash) *Block {

	h := &Header{
		Version:       1,
		PrevBlockHash: prevblockhash,
		Timestamp:     time.Now().UnixNano(),
		Height:        height,
	}
	t := Transaction{
		Data: []byte("hello world"),
	}
	return NewBlock(h, []Transaction{t})
}

func RandomBlockWithSignature(t *testing.T, height uint32, prevblockhash types.Hash) *Block {
	privatekey := crypto.GeneratePrivateKey()
	block := RandomBlock(height, prevblockhash)
	assert.Nil(t, block.Sign(privatekey))
	return block
}

func TestBlock_Hash(t *testing.T) {
	block := RandomBlock(0, types.Hash{})
	fmt.Println(block.Hash(BlockHasher{}))
}
