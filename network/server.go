package network

import (
	"bytes"
	"github.com/go-kit/log"
	"github.com/sirupsen/logrus"
	"github.com/virtue186/xchain/core"
	"github.com/virtue186/xchain/crypto"
	"github.com/virtue186/xchain/types"
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
	chain, err := core.NewBlockChain(genesisBlock())
	if err != nil {
		return nil, err
	}

	s := &Server{
		BlockChain:  chain,
		ServerOpts:  opts,
		blocktime:   opts.BlockTime,
		memPool:     NewTxPool(),
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
	}
	return nil
}

func (s *Server) ProcessTransaction(tx *core.Transaction) error {

	hash := tx.Hash(core.TxHasher{})
	if s.memPool.Has(hash) {
		logrus.WithFields(logrus.Fields{
			"hash": hash,
		}).Info("transaction already exists")
		return nil
	}

	if err := tx.Verify(); err != nil {
		return err
	}
	tx.SetFirstSeen(time.Now().UnixNano())

	s.Logger.Log("msg", "add new transaction to pool",
		"hash", hash,
		"mempool length", s.memPool.Len(),
	)

	go s.broadcastTx(tx)

	return s.memPool.Add(tx)

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

	ticker := time.NewTicker(s.blocktime)
	s.Logger.Log("msg", "validator loop started", "blocktime", s.blocktime)
	for {
		<-ticker.C
		s.CreateNewBlock()
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
	block, err := core.NewBlockFromPreHeader(currentHeader, nil)
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

	return nil
}

func genesisBlock() *core.Block {
	header := &core.Header{
		Version:       1,
		PrevBlockHash: types.Hash{},
		Height:        0,
		Timestamp:     time.Now().UnixNano(),
		DataHash:      types.Hash{},
	}
	return core.NewBlock(header, nil)

}

func (s *Server) broadcast(payload []byte) error {
	for _, transport := range s.Transports {
		if err := transport.Broadcast(payload); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) broadcastTx(tx *core.Transaction) error {
	buf := &bytes.Buffer{}
	if err := tx.Encode(core.NewGobTxEncoder(buf)); err != nil {
		return err
	}
	msg := NewMessage(MessageTypeTx, buf.Bytes())
	return s.broadcast(msg.Bytes())
}
