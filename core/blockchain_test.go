package core

import (
	"github.com/go-kit/log"
	"github.com/stretchr/testify/assert"
	"github.com/virtue186/xchain/types"
	"testing"
)

func TestAddBlock(t *testing.T) {
	bc := newBlockchainWithGenesis(t)

	lenBlocks := 1000
	for i := 0; i < lenBlocks; i++ {
		block := randomBlock(t, uint32(i+1), getPrevBlockHash(t, bc, uint32(i+1)))
		assert.Nil(t, bc.AddBlock(block))
	}

	assert.Equal(t, bc.Height(), uint32(lenBlocks))
	assert.Equal(t, len(bc.headers), lenBlocks+1)
	assert.NotNil(t, bc.AddBlock(randomBlock(t, 89, types.Hash{})))
}

func TestNewBlockchain(t *testing.T) {
	bc := newBlockchainWithGenesis(t)
	assert.NotNil(t, bc.validator)
	assert.Equal(t, bc.Height(), uint32(0))
}

func TestGetHeader(t *testing.T) {
	bc := newBlockchainWithGenesis(t)
	lenBlocks := 1000

	for i := 0; i < lenBlocks; i++ {
		block := randomBlock(t, uint32(i+1), getPrevBlockHash(t, bc, uint32(i+1)))
		assert.Nil(t, bc.AddBlock(block))
		header, err := bc.GetHeader(block.Height)
		assert.Nil(t, err)
		assert.Equal(t, header, block.Header)
	}
}

func TestAddBlockToHigh(t *testing.T) {
	bc := newBlockchainWithGenesis(t)

	assert.Nil(t, bc.AddBlock(randomBlock(t, 1, getPrevBlockHash(t, bc, uint32(1)))))
	assert.NotNil(t, bc.AddBlock(randomBlock(t, 3, types.Hash{})))
}

func newBlockchainWithGenesis(t *testing.T) *BlockChain {
	bc, err := NewBlockChain(log.NewNopLogger(), randomBlock(t, 0, types.Hash{}))
	assert.Nil(t, err)

	return bc
}

func getPrevBlockHash(t *testing.T, bc *BlockChain, height uint32) types.Hash {
	prevHeader, err := bc.GetHeader(height - 1)
	assert.Nil(t, err)
	return BlockHasher{}.Hash(prevHeader)
}
