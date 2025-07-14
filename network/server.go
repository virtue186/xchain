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

var defaultBlockTime = time.Second * 5

type ServerOpts struct {
	ID            string
	Logger        log.Logger
	RPCDecodeFunc RPCDecodeFunc
	RPCProcessor  RPCProcessor
	Transports    []Transport
	PrivateKey    *crypto.PrivateKey
	BlockTime     time.Duration
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
	if opts.RPCDecodeFunc == nil {
		opts.RPCDecodeFunc = DefaultRPCDecodeFunc
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
	go s.broadBlock(block)
	return nil
}

func (s *Server) Start() {
	s.InitTransports()
free:
	for {
		select {
		case rpc := <-s.rpcCh:
			msg, err := s.RPCDecodeFunc(rpc)
			if err != nil {
				s.Logger.Log("err", err)
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

func (s *Server) validatorLoop() {
	ticker := time.NewTicker(s.BlockTime)

	s.Logger.Log("msg", "Starting validator loop", "blockTime", s.BlockTime)

	for {
		fmt.Println("creating new block")

		if err := s.CreateNewBlock(); err != nil {
			s.Logger.Log("create block error", err)
		}

		<-ticker.C
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
	err := core.NewGobBlockEncoder(buf).Encode(block)
	if err != nil {
		return err
	}

	message := NewMessage(MessageTypeBlock, buf.Bytes())

	return s.broadcast(message.Bytes())
}

func (s *Server) broadcastTx(tx *core.Transaction) error {
	buf := &bytes.Buffer{}
	if err := tx.Encode(core.NewGobTxEncoder(buf)); err != nil {
		return err
	}
	msg := NewMessage(MessageTypeTx, buf.Bytes())
	return s.broadcast(msg.Bytes())
}
