package node

import (
	"bytes"
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
	server        *network.Server
}

func NewChainService(bc *core.BlockChain, txPool *network.TxPool, logger log.Logger, txb chan<- *core.Transaction, server *network.Server) *ChainService {
	return &ChainService{
		blockChain:    bc,
		txPool:        txPool,
		logger:        logger,
		txBroadcaster: txb,
		server:        server,
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
	case *network.GetStatusMessage:
		return s.handleGetStatusMessage(msg.From, t)
	case *network.StatusMessage:
		return s.handleStatusMessage(msg.From, t)
	case *network.GetBlocksMessage:
		return s.handleGetBlocksMessage(msg.From, t)
	case *network.BlocksMessage:
		return s.handleBlocksMessage(msg.From, t)
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

// handleGetStatusMessage 当收到状态请求时，回复自己的状态
func (s *ChainService) handleGetStatusMessage(from network.NetAddr, data *network.GetStatusMessage) error {
	s.logger.Log("msg", "received get_status message", "from", from)
	statusMsg := &network.StatusMessage{
		ID:            s.server.ID,
		CurrentHeight: s.blockChain.Height(),
	}

	buf := new(bytes.Buffer)
	if err := s.server.Encoder.Encode(buf, statusMsg); err != nil {
		return err
	}
	msg := network.NewMessage(network.MessageTypeStatus, buf.Bytes())
	finalBuf := new(bytes.Buffer)
	if err := s.server.Encoder.Encode(finalBuf, msg); err != nil {
		return err
	}

	s.logger.Log("msg", "sending status message", "to", from, "height", statusMsg.CurrentHeight)
	return s.server.SendMessage(from, finalBuf.Bytes())
}

// handleStatusMessage 收到对方的状态后，决策是否需要同步
func (s *ChainService) handleStatusMessage(from network.NetAddr, data *network.StatusMessage) error {
	s.logger.Log("msg", "received status message", "from", from, "peerHeight", data.CurrentHeight)

	myHeight := s.blockChain.Height()
	if data.CurrentHeight > myHeight {
		s.logger.Log("msg", "starting sync with peer", "peerHeight", data.CurrentHeight, "myHeight", myHeight)
		// 直接调用封装好的方法来请求区块
		return s.requestBlocks(from, myHeight+1)
	}

	s.logger.Log("msg", "no sync needed", "from", from)
	return nil
}

// handleGetBlocksMessage 收到区块请求，从数据库读取并发送回去
func (s *ChainService) handleGetBlocksMessage(from network.NetAddr, data *network.GetBlocksMessage) error {
	s.logger.Log("msg", "received get_blocks message", "from", from, "start", data.From)

	// 一次最多发送100个区块
	blocks, err := s.blockChain.GetBlocks(data.From, 100)
	if err != nil {
		return err
	}

	if len(blocks) == 0 {
		s.logger.Log("msg", "no blocks to send", "from", from, "start", data.From)
		return nil
	}

	s.logger.Log("msg", "sending blocks", "to", from, "count", len(blocks))

	blocksData := make([][]byte, len(blocks))
	for i, block := range blocks {
		buf := new(bytes.Buffer)
		if err := block.Encode(buf, core.GOBEncoder[*core.Block]{}); err != nil { // 假设 Block 有 Encode 方法
			return err
		}
		blocksData[i] = buf.Bytes()
	}

	blocksMsg := &network.BlocksMessage{
		Blocks: blocksData,
	}

	buf := new(bytes.Buffer)
	if err := s.server.Encoder.Encode(buf, blocksMsg); err != nil {
		return err
	}
	msg := network.NewMessage(network.MessageTypeBlocks, buf.Bytes())
	finalBuf := new(bytes.Buffer)
	if err := s.server.Encoder.Encode(finalBuf, msg); err != nil {
		return err
	}

	return s.server.SendMessage(from, finalBuf.Bytes())
}

// handleBlocksMessage 收到区块数据，进行验证并添加到本地区块链
func (s *ChainService) handleBlocksMessage(from network.NetAddr, data *network.BlocksMessage) error {
	s.logger.Log("msg", "received blocks message", "from", from, "count", len(data.Blocks))

	if len(data.Blocks) == 0 {
		s.logger.Log("msg", "peer sent empty blocks message, sync might be complete", "from", from)
		return nil
	}

	for _, blockData := range data.Blocks {
		block, err := core.DecodeBlock(blockData)
		if err != nil {
			return err
		}

		if err := s.blockChain.AddBlock(block); err != nil {
			s.logger.Log("msg", "failed to add synced block, stopping sync with this peer.", "err", err, "height", block.Height)
			// 如果一个区块验证失败，则立即停止处理后续区块，并停止向该节点同步
			return err
		}
	}

	// --- 完善后的持续同步逻辑 ---

	// 定义每批次请求的最大数量 (与 handleGetBlocksMessage 中的一致)
	const maxBlocksPerRequest = 100

	// 如果我们收到的区块数量等于我们一次请求的最大数量，
	// 那么有很高的概率对方还有更多的区块。
	if len(data.Blocks) == maxBlocksPerRequest {
		s.logger.Log("msg", "finished processing batch, requesting next batch", "from", from)

		// 立即请求下一批区块，无缝衔接
		return s.requestBlocks(from, s.blockChain.Height()+1)
	}

	// 如果收到的区块数量小于最大值，说明我们很可能已经追上了对方的链。
	// 同步流程自然结束。
	s.logger.Log("msg", "sync likely complete with peer", "from", from, "newHeight", s.blockChain.Height())
	return nil
}

func (s *ChainService) requestBlocks(to network.NetAddr, fromHeight uint32) error {
	getBlocksMsg := &network.GetBlocksMessage{
		From: fromHeight,
	}

	buf := new(bytes.Buffer)
	if err := s.server.Encoder.Encode(buf, getBlocksMsg); err != nil {
		return err
	}
	msg := network.NewMessage(network.MessageTypeGetBlocks, buf.Bytes())
	finalBuf := new(bytes.Buffer)
	if err := s.server.Encoder.Encode(finalBuf, msg); err != nil {
		return err
	}

	s.logger.Log("msg", "sending get_blocks message", "to", to, "fromHeight", fromHeight)
	return s.server.SendMessage(to, finalBuf.Bytes())
}
