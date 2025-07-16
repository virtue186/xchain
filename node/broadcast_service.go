package node

import (
	"bytes"
	"github.com/go-kit/log"
	"github.com/virtue186/xchain/core"
	"github.com/virtue186/xchain/network"
)

type BroadcastService struct {
	logger  log.Logger
	server  *network.Server   // 持有Server引用，以调用其Broadcast方法
	encoder core.Encoder[any] // 持有统一的编码器

	blockChan chan *core.Block
	txChan    chan *core.Transaction
}

func NewBroadcastService(l log.Logger, s *network.Server, e core.Encoder[any]) *BroadcastService {
	return &BroadcastService{
		logger:    l,
		server:    s,
		encoder:   e,
		blockChan: make(chan *core.Block, 10),       // 使用带缓冲的channel
		txChan:    make(chan *core.Transaction, 10), // 使用带缓冲的channel
	}
}

// BlockBroadcastChan 返回一个只写的channel，供外部发送区块
func (bs *BroadcastService) BlockBroadcastChan() chan<- *core.Block {
	return bs.blockChan
}

// TxBroadcastChan 返回一个只写的channel，供外部发送交易
func (bs *BroadcastService) TxBroadcastChan() chan<- *core.Transaction {
	return bs.txChan
}

// Start 启动广播服务的主循环
func (bs *BroadcastService) Start() {
	bs.logger.Log("msg", "starting broadcast service")
	for {
		select {
		case block := <-bs.blockChan:
			if err := bs.broadcastBlock(block); err != nil {
				bs.logger.Log("msg", "failed to broadcast block", "err", err)
			}
		case tx := <-bs.txChan:
			if err := bs.broadcastTransaction(tx); err != nil {
				bs.logger.Log("msg", "failed to broadcast transaction", "err", err)
			}
		}
	}
}

func (bs *BroadcastService) broadcastBlock(block *core.Block) error {
	bs.logger.Log("msg", "broadcasting new block", "hash", block.Hash(core.BlockHasher{}))
	return bs.broadcast(network.MessageTypeBlock, block)
}

func (bs *BroadcastService) broadcastTransaction(tx *core.Transaction) error {
	bs.logger.Log("msg", "broadcasting new transaction", "hash", tx.Hash(core.TxHasher{}))
	return bs.broadcast(network.MessageTypeTx, tx)
}

// broadcast 是一个通用的辅助函数，负责编码和广播
func (bs *BroadcastService) broadcast(msgType network.MessageType, data any) error {
	buf := &bytes.Buffer{}
	if err := bs.encoder.Encode(buf, data); err != nil {
		return err
	}

	msg := network.NewMessage(msgType, buf.Bytes())

	finalBuf := &bytes.Buffer{}
	if err := bs.encoder.Encode(finalBuf, msg); err != nil {
		return err
	}

	return bs.server.Broadcast(finalBuf.Bytes())
}
