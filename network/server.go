package network

import (
	"bytes"
	"fmt"
	"github.com/go-kit/log"
	"github.com/virtue186/xchain/core"
	"github.com/virtue186/xchain/crypto"
	"os"
	"time"
)

var defaultBlockTime = time.Second * 15

type ServerOpts struct {
	ID           string
	Logger       log.Logger
	RPCProcessor RPCProcessor
	Transports   []Transport
	PrivateKey   *crypto.PrivateKey
	BlockTime    time.Duration
	// 新增：统一的编码器和解码器
	Encoder core.Encoder[any]
	Decoder core.Decoder[any]
}

type Server struct {
	ServerOpts
	blocktime   time.Duration
	memPool     *TxPool
	IsValidator bool
	BlockChain  *core.BlockChain
	rpcCh       chan RPC
	quitCh      chan struct{}
}

func NewServer(opts ServerOpts) (*Server, error) {
	if opts.BlockTime == time.Duration(0) {
		opts.BlockTime = defaultBlockTime
	}
	if opts.Encoder == nil {
		opts.Encoder = core.GOBEncoder[any]{}
	}
	if opts.Decoder == nil {
		opts.Decoder = core.GOBDecoder[any]{}
	}
	if opts.Logger == nil {
		opts.Logger = log.NewLogfmtLogger(os.Stderr)
		opts.Logger = log.With(opts.Logger, "ID", opts.ID)
	}
	chain, err := core.NewBlockChain(opts.Logger, genesisBlock())
	if err != nil {
		return nil, err
	}

	s := &Server{
		BlockChain:  chain,
		ServerOpts:  opts,
		blocktime:   opts.BlockTime,
		memPool:     NewTxPool(1000),
		IsValidator: opts.PrivateKey != nil,
		rpcCh:       make(chan RPC),
		quitCh:      make(chan struct{}, 1),
	}

	if s.RPCProcessor == nil {
		s.RPCProcessor = s
	}
	if s.IsValidator {
		go s.validatorLoop()
	}

	return s, nil
}

func (s *Server) ProcessMessage(message *DecodedMessage) error {

	switch t := message.Data.(type) {
	case *core.Transaction:
		return s.ProcessTransaction(t)
	case *core.Block:
		return s.ProcessBlock(t)
	}
	return nil
}

func (s *Server) ProcessTransaction(tx *core.Transaction) error {

	hash := tx.Hash(core.TxHasher{})
	if s.memPool.Contains(hash) {
		return nil
	}

	if err := tx.Verify(); err != nil {
		return err
	}
	tx.SetFirstSeen(time.Now().UnixNano())

	s.Logger.Log(
		"msg", "adding new tx to mempool",
		"hash", hash,
		"mempoolPending", s.memPool.PendingCount(),
	)

	go s.broadcastTx(tx)
	s.memPool.Add(tx)
	return nil
}

func (s *Server) ProcessBlock(block *core.Block) error {
	err := s.BlockChain.AddBlock(block)
	if err != nil {
		return err
	}
	//go s.broadBlock(block)
	s.memPool.Flush(block.Transactions)
	return nil
}

func (s *Server) Start() {
	s.InitTransports()
free:
	for {
		select {
		case rpc := <-s.rpcCh:
			// 使用注入的解码器进行解码
			msg, err := s.decodeMessage(rpc.Message)
			if err != nil {
				s.Logger.Log("err", err)
				continue // 如果解码失败，继续处理下一条消息
			}
			err = s.RPCProcessor.ProcessMessage(msg)
			if err != nil {
				s.Logger.Log("err", err)
			}

		case <-s.quitCh:
			break free
		}
	}
	s.Logger.Log("server is shutting down")
}

func (s *Server) decodeMessage(msg *Message) (*DecodedMessage, error) {
	// 这里的 Decoder 应该是 Server 自己的，用于解码 Message.Data
	// 为了简单起见，我们暂时直接创建
	decoder := core.GOBDecoder[any]{}

	switch msg.Header {
	case MessageTypeTx:
		tx := new(core.Transaction)
		if err := decoder.Decode(bytes.NewReader(msg.Data), tx); err != nil {
			return nil, fmt.Errorf("decode transaction error: %w", err)
		}
		return &DecodedMessage{Data: tx}, nil
	case MessageTypeBlock:
		b := new(core.Block)
		if err := decoder.Decode(bytes.NewReader(msg.Data), b); err != nil {
			return nil, fmt.Errorf("decode block error: %w", err)
		}
		return &DecodedMessage{Data: b}, nil
	default:
		return nil, fmt.Errorf("unknown message header: %v", msg.Header)
	}
}

func (s *Server) validatorLoop() {
	ticker := time.NewTicker(s.BlockTime)

	s.Logger.Log("msg", "Starting validator loop", "blockTime", s.BlockTime)

	for {
		<-ticker.C
		fmt.Println("creating new block")
		if err := s.CreateNewBlock(); err != nil {
			s.Logger.Log("create block error", err)
		}

	}
}

func (s *Server) InitTransports() {

	for _, transport := range s.Transports {
		go func(transport Transport) {
			for rpc := range transport.Consume() {
				s.rpcCh <- rpc
			}

		}(transport)

	}

}

func (s *Server) CreateNewBlock() error {

	currentHeader, err := s.BlockChain.GetHeader(s.BlockChain.Height())

	if err != nil {
		return err
	}

	txx := s.memPool.Pending()

	block, err := core.NewBlockFromPreHeader(currentHeader, txx)
	if err != nil {
		return err
	}
	err = block.Sign(*s.PrivateKey)
	if err != nil {
		return err
	}

	err = s.BlockChain.AddBlock(block)
	if err != nil {
		return err
	}
	s.memPool.ClearPending()
	go s.broadBlock(block)

	return nil
}

func genesisBlock() *core.Block {
	dataHash, err := core.CalculateDataHash(nil)
	if err != nil {
		// This panic is appropriate because if hashing nil fails, the program is in an unrecoverable state.
		panic(fmt.Sprintf("failed to create genesis block data hash: %v", err))
	}
	header := &core.Header{
		Version:   1,
		Height:    0,
		Timestamp: 000000,
		DataHash:  dataHash,
	}
	block, err := core.NewBlock(header, nil)
	if err != nil {
		fmt.Errorf("error creating genesis block: %v", err)
	}
	return block

}

func (s *Server) broadcast(payload []byte) error {
	for _, transport := range s.Transports {
		if err := transport.Broadcast(payload); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) broadBlock(block *core.Block) error {
	buf := &bytes.Buffer{}
	// 直接调用注入的编码器
	if err := s.Encoder.Encode(buf, block); err != nil {
		return err
	}

	message := NewMessage(MessageTypeBlock, buf.Bytes())

	return s.broadcast(message.Bytes())
}

func (s *Server) broadcastTx(tx *core.Transaction) error {
	// 1. 将 Transaction 对象编码成字节流
	txBuf := &bytes.Buffer{}
	if err := s.Encoder.Encode(txBuf, tx); err != nil {
		return fmt.Errorf("failed to encode transaction for broadcast: %w", err)
	}

	// 2. 创建 Message，将编码后的交易字节流放入 Data
	msg := NewMessage(MessageTypeTx, txBuf.Bytes())

	// 3. 将完整的 Message 对象编码后广播出去
	finalBuf := &bytes.Buffer{}
	if err := s.Encoder.Encode(finalBuf, msg); err != nil {
		return fmt.Errorf("failed to encode message for broadcast: %w", err)
	}

	return s.broadcast(finalBuf.Bytes())
}
