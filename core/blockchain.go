package core

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"sync"
)

type BlockChain struct {
	store     Storage
	headers   []*Header
	validator Validator

	lock sync.RWMutex
}

func NewBlockChain(genesis *Block) (*BlockChain, error) {
	bc := &BlockChain{
		headers: []*Header{},
		store:   NewMemoryStorage(),
	}
	bc.validator = NewBlockValidator(bc)
	err := bc.AddBlockWithoutValidation(genesis)
	return bc, err
}

func (bc *BlockChain) AddBlock(b *Block) error {
	bc.lock.Lock()
	defer bc.lock.Unlock()

	if err := bc.validator.ValidateBlock(b); err != nil {
		return err
	}
	return bc.AddBlockWithoutValidation(b)
}

func (bc *BlockChain) SetValidator(v Validator) {
	bc.validator = v
}

func (bc *BlockChain) Height() uint32 {
	bc.lock.RLock()
	defer bc.lock.RUnlock()
	return uint32(len(bc.headers) - 1)
}

func (bc *BlockChain) AddBlockWithoutValidation(b *Block) error {
	logrus.WithFields(logrus.Fields{
		"height": b.Height,
		"hash":   b.Hash(BlockHasher{}),
	}).Info("add block")

	bc.headers = append(bc.headers, b.Header)
	return bc.store.Put(b)
}

func (bc *BlockChain) GetHeader(height uint32) (*Header, error) {
	bc.lock.RLock()
	defer bc.lock.RUnlock()

	if height > uint32(len(bc.headers)-1) {
		return nil, fmt.Errorf("height %d out of range %d", height, len(bc.headers)-1)
	}
	return bc.headers[height], nil

}
