package node

import (
	"fmt"
	"github.com/go-kit/log"
	"github.com/virtue186/xchain/core"
	"github.com/virtue186/xchain/network"
	"time"
)

type ChainService struct {
	logger        log.Logger
	blockChain    *core.BlockChain
	txPool        *network.TxPool
	txBroadcaster chan<- *core.Transaction
}

func NewChainService(bc *core.BlockChain, txPool *network.TxPool, logger log.Logger, txb chan<- *core.Transaction) *ChainService {
	return &ChainService{
		blockChain:    bc,
		txPool:        txPool,
		logger:        logger,
		txBroadcaster: txb,
	}
}

func (s *ChainService) OnPeer(p *network.TCPPeer) error {
	// 在这里可以实现未来的区块同步逻辑，例如向新节点请求其状态
	return nil
}

func (s *ChainService) ProcessMessage(msg *network.DecodedMessage) error {
	switch t := msg.Data.(type) {
	case *core.Transaction:
		return s.ProcessTransaction(t)
	case *core.Block:
		return s.ProcessBlock(t)
	default:
		return fmt.Errorf("chain service received unknown message type: %T", t)
	}
}

func (s *ChainService) ProcessTransaction(tx *core.Transaction) error {

	hash := tx.Hash(core.TxHasher{})
	if s.txPool.Contains(hash) {
		return nil
	}

	if err := tx.Verify(); err != nil {
		return err
	}
	tx.SetFirstSeen(time.Now().UnixNano())

	s.logger.Log(
		"msg", "adding new tx to mempool",
		"hash", hash,
		"mempoolPending", s.txPool.PendingCount(),
	)
	go func() {
		s.txBroadcaster <- tx
	}()
	s.txPool.Add(tx)
	return nil
}

func (s *ChainService) ProcessBlock(block *core.Block) error {
	if err := s.blockChain.AddBlock(block); err != nil {
		s.logger.Log("msg", "failed to add block", "error", err, "height", block.Height)
		return err
	}

	s.txPool.Flush(block.Transactions)
	s.logger.Log("msg", "flushed mempool", "count", len(block.Transactions))

	// TODO  未实现广播收到的区块

	return nil
}
