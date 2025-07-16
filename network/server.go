package network

import (
	"bytes"
	"fmt"
	"github.com/go-kit/log"
	"github.com/virtue186/xchain/core"
	"os"
)

type ServerOpts struct {
	ID           string
	Logger       log.Logger
	RPCProcessor RPCProcessor // 依赖注入！
	Transports   []Transport
	Encoder      core.Encoder[any] // 新增：统一的编码器
	Decoder      core.Decoder[any] // 新增：统一的解码器
}

type Server struct {
	ServerOpts
	rpcCh  chan RPC
	quitCh chan struct{}
}

func NewServer(opts ServerOpts) *Server {
	if opts.Logger == nil {
		opts.Logger = log.NewLogfmtLogger(os.Stderr)
		opts.Logger = log.With(opts.Logger, "ID", opts.ID)
	}
	if opts.Encoder == nil {
		opts.Encoder = core.GOBEncoder[any]{}
	}
	if opts.Decoder == nil {
		opts.Decoder = core.GOBDecoder[any]{}
	}

	s := &Server{
		ServerOpts: opts,
		rpcCh:      make(chan RPC),
		quitCh:     make(chan struct{}, 1),
	}
	if s.RPCProcessor == nil {
		s.RPCProcessor = &NOPRPCProcessor{}
	}

	return s
}

func (s *Server) Start() {
	s.InitTransports()
free:
	for {
		select {
		case rpc := <-s.rpcCh:
			msg := rpc.Message
			from := rpc.From

			decodedMsg, err := s.decodeMessageData(msg, from)
			if err != nil {
				s.Logger.Log("msg", "failed to decode message data", "err", err, "from", from)
				continue
			}
			if err := s.RPCProcessor.ProcessMessage(decodedMsg); err != nil {
				s.Logger.Log("msg", "failed to process message", "err", err, "from", from)
			}

		case <-s.quitCh:
			break free
		}
	}
	s.Logger.Log("msg", "server is shutting down")
}

func (s *Server) InitTransports() {
	for _, transport := range s.Transports {
		go func(tr Transport) {
			for rpc := range tr.Consume() {
				s.rpcCh <- rpc
			}
		}(transport)
	}
}

func (s *Server) Broadcast(payload []byte) error {
	for _, tr := range s.Transports {
		if err := tr.Broadcast(payload); err != nil {
			return err
		}
	}
	return nil
}

// decodeMessageData 是 Server 的一个辅助方法，负责解码 Message.Data
func (s *Server) decodeMessageData(msg *Message, from NetAddr) (*DecodedMessage, error) {
	decodedMsg := &DecodedMessage{From: from}

	switch msg.Header {
	case MessageTypeTx:
		tx := new(core.Transaction)
		if err := s.Decoder.Decode(bytes.NewReader(msg.Data), tx); err != nil {
			return nil, fmt.Errorf("decode transaction error: %w", err)
		}
		decodedMsg.Data = tx

	case MessageTypeBlock:
		b := new(core.Block)
		if err := s.Decoder.Decode(bytes.NewReader(msg.Data), b); err != nil {
			return nil, fmt.Errorf("decode block error: %w", err)
		}
		decodedMsg.Data = b

	default:
		return nil, fmt.Errorf("unknown message header: %v", msg.Header)
	}

	return decodedMsg, nil
}
