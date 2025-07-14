package core

import (
	"fmt"
	"github.com/go-kit/log"
	"sync"
)

type BlockChain struct {
	logger        log.Logger
	store         Storage
	headers       []*Header
	validator     Validator
	lock          sync.RWMutex
	contractState *State
}

func NewBlockChain(log log.Logger, genesis *Block) (*BlockChain, error) {
	bc := &BlockChain{
		contractState: NewState(),
		headers:       []*Header{},
		store:         NewMemoryStorage(),
		logger:        log,
	}
	bc.validator = NewBlockValidator(bc)
	err := bc.AddBlockWithoutValidation(genesis)
	return bc, err
}

func (bc *BlockChain) AddBlock(b *Block) error {
	if err := bc.validator.ValidateBlock(b); err != nil {
		return err
	}
	for _, tx := range b.Transactions {
		bc.logger.Log("msg", "executing code", "len", len(tx.Data), "hash", tx.Hash(&TxHasher{}))
		vm := NewVm(tx.Data, NewState())
		err := vm.Run()
		if err != nil {
			return err
		}
		fmt.Printf("State: +%v\n", vm.contractState)
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
	bc.logger.Log(
		"msg", "add block",
		"hash", b.Hash(BlockHasher{}),
		"height", b.Height,
		"transaction", len(b.Transactions),
	)

	bc.headers = append(bc.headers, b.Header)
	return bc.store.Put(b)
}

func (bc *BlockChain) GetHeader(height uint32) (*Header, error) {
	if height > bc.Height() {
		return nil, fmt.Errorf("given height (%d) too high", height)
	}

	bc.lock.Lock()
	defer bc.lock.Unlock()

	return bc.headers[height], nil
}
