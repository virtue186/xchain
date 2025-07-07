package network

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/virtue186/xchain/core"
	"github.com/virtue186/xchain/crypto"
	"time"
)

var defaultBlockTime = time.Second * 5

type ServerOpts struct {
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
	rpcCh       chan RPC
	quitCh      chan struct{}
}

func NewServer(opts ServerOpts) *Server {
	if opts.BlockTime == time.Duration(0) {
		opts.BlockTime = defaultBlockTime
	}
	if opts.RPCDecodeFunc == nil {
		opts.RPCDecodeFunc = DefaultRPCDecodeFunc
	}
	s := &Server{
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

	return s
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

	logrus.WithFields(logrus.Fields{
		"hash":           tx.Hash(core.TxHasher{}),
		"mempool length": s.memPool.Len(),
	}).Info("add new transaction to pool")

	go s.broadcastTx(tx)

	return s.memPool.Add(tx)

}

func (s *Server) Start() {
	s.InitTransports()
	ticker := time.NewTicker(s.blocktime)
free:
	for {
		select {
		case rpc := <-s.rpcCh:
			msg, err := s.RPCDecodeFunc(rpc)
			if err != nil {
				logrus.Errorf("handle rpc error: %v", err)
			}
			err2 := s.RPCProcessor.ProcessMessage(msg)
			if err != nil {
				logrus.Errorf("handle rpc error: %v", err2)
			}

		case <-s.quitCh:
			break free
		case <-ticker.C:
			if s.IsValidator {
				s.CreateBlock()
			}
		}
	}
	fmt.Println("server stopped")
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

func (s *Server) CreateBlock() error {
	fmt.Println("create a new block")
	return nil
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
