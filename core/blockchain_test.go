package core

import (
	"github.com/stretchr/testify/assert"
	"github.com/virtue186/xchain/types"
	"testing"
)

func NewBlockWithoutValidator(t *testing.T) *BlockChain {
	bc, err := NewBlockChain(RandomBlock(0, types.Hash{}))
	assert.Nil(t, err)
	return bc
}

func TestBlockChain(t *testing.T) {
	bc, err := NewBlockChain(RandomBlock(0, types.Hash{}))
	assert.Nil(t, err)
	assert.NotNil(t, bc.validator)
	assert.Equal(t, bc.Height(), uint32(0))
	println(bc.Height())
}

func TestAddBlock(t *testing.T) {
	bc := NewBlockWithoutValidator(t)
	for i := 0; i < 1000; i++ {
		block := RandomBlockWithSignature(t, uint32(i+1), getPrevBlockHash(t, bc, uint32(i+1)))
		err := bc.AddBlock(block)
		assert.Nil(t, err)
	}
	assert.Equal(t, bc.Height(), uint32(1000))
	assert.Equal(t, len(bc.headers), 1001)
	assert.Nil(t, bc.AddBlock(RandomBlock(18, types.Hash{})))
}

func getPrevBlockHash(t *testing.T, bc *BlockChain, height uint32) types.Hash {
	prevHeader, err := bc.GetHeader(height - 1)
	assert.Nil(t, err)
	return BlockHasher{}.Hash(prevHeader)
}
