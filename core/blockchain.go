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
	State         *State
}

func NewBlockChain(log log.Logger, storage Storage, genesis *Block) (*BlockChain, error) {
	bc := &BlockChain{
		contractState: NewState(storage),
		headers:       []*Header{},
		store:         storage,
		logger:        log,
		State:         NewState(storage),
	}
	bc.validator = NewBlockValidator(bc)
	// 从数据库加载现有的区块头
	if err := bc.loadHeaders(); err != nil {
		// 如果加载失败（比如数据库是空的），则添加创世块
		if err.Error() == "database is empty" {
			bc.logger.Log("msg", "database empty, adding genesis block")
			return bc, bc.AddBlockWithoutValidation(genesis)
		}
		return nil, err
	}

	return bc, nil
}

// loadHeaders 从数据库加载所有区块头到内存中
func (bc *BlockChain) loadHeaders() error {
	// 我们需要一个方法来知道数据库中的最高高度，这里我们先用一个迭代的方式
	// 更好的方式是单独存储一个 "LATEST_HEIGHT" 的键值对
	var height uint32 = 0
	for {
		hash, err := bc.store.GetBlockHashByHeight(height)
		if err != nil {
			// 如果在高度0都找不到哈希，说明数据库是空的
			if height == 0 {
				return fmt.Errorf("database is empty")
			}
			// 否则，说明已到达最高高度
			break
		}

		block, err := bc.store.GetBlockByHash(hash)
		if err != nil {
			return err
		}

		bc.headers = append(bc.headers, block.Header)
		height++
	}

	bc.logger.Log("msg", "loaded headers from disk", "count", len(bc.headers))
	return nil
}

func (bc *BlockChain) GetBlocks(fromHeight uint32, count int) ([]*Block, error) {
	bc.lock.RLock()
	defer bc.lock.RUnlock()

	// 确保请求范围有效
	if fromHeight > bc.Height() {
		return nil, nil // 没有新区块可提供
	}

	blocks := make([]*Block, 0, count)
	// 最多提供到链的最高点
	for i := 0; i < count && fromHeight+uint32(i) <= bc.Height(); i++ {
		currentHeight := fromHeight + uint32(i)
		hash, err := bc.store.GetBlockHashByHeight(currentHeight)
		if err != nil {
			return nil, err
		}
		block, err := bc.store.GetBlockByHash(hash)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, block)
	}

	return blocks, nil
}

func (bc *BlockChain) AddBlock(b *Block) error {
	if err := bc.validator.ValidateBlock(b); err != nil {
		return err
	}
	if err := bc.applyBlock(b); err != nil {
		// 如果交易应用失败，这是一个严重的共识错误，不应添加此区块
		return fmt.Errorf("failed to apply block: %w", err)
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
	return bc.store.PutBlock(b)
}

func (bc *BlockChain) GetHeader(height uint32) (*Header, error) {
	if height > bc.Height() {
		return nil, fmt.Errorf("given height (%d) too high", height)
	}

	bc.lock.Lock()
	defer bc.lock.Unlock()

	return bc.headers[height], nil
}

func (bc *BlockChain) applyBlock(b *Block) error {
	for _, tx := range b.Transactions {
		if err := bc.applyTransaction(tx); err != nil {
			return err
		}
	}
	return nil
}

// applyTransaction 是状态转换的核心函数
func (bc *BlockChain) applyTransaction(tx *Transaction) error {
	senderAddr := tx.From.Address()

	// 1. 获取发送方和接收方的账户状态
	senderState, err := bc.State.Get(senderAddr)
	if err != nil {
		return err
	}

	receiverState, err := bc.State.Get(tx.To)
	if err != nil {
		return err
	}

	// 2. 验证交易
	// 2.1 验证 Nonce
	if tx.Nonce != senderState.Nonce {
		return fmt.Errorf("invalid nonce. expected %d, got %d", senderState.Nonce, tx.Nonce)
	}
	// 2.2 验证余额
	if senderState.Balance < tx.Value {
		return fmt.Errorf("insufficient balance. have %d, want %d", senderState.Balance, tx.Value)
	}

	// 3. 执行状态转换
	senderState.Nonce++
	senderState.Balance -= tx.Value
	receiverState.Balance += tx.Value

	// 4. 将更新后的状态写回数据库
	if err := bc.State.Put(senderAddr, senderState); err != nil {
		return err
	}
	if err := bc.State.Put(tx.To, receiverState); err != nil {
		return err
	}

	bc.logger.Log("msg", "transaction applied", "from", senderAddr, "to", tx.To, "value", tx.Value)

	return nil
}
