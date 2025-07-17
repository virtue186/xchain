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
		opts.Encoder = core.JSONEncoder[any]{}
	}
	if opts.Decoder == nil {
		opts.Decoder = core.JSONDecoder[any]{}
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

	case MessageTypeGetStatus:
		getStatusMsg := new(GetStatusMessage)
		if err := s.Decoder.Decode(bytes.NewReader(msg.Data), getStatusMsg); err != nil {
			return nil, fmt.Errorf("failed to decode getstatus message: %w", err)
		}
		decodedMsg.Data = getStatusMsg

	case MessageTypeStatus:
		statusMsg := new(StatusMessage)
		if err := s.Decoder.Decode(bytes.NewReader(msg.Data), statusMsg); err != nil {
			return nil, fmt.Errorf("failed to decode status message: %w", err)
		}
		decodedMsg.Data = statusMsg

	case MessageTypeGetBlocks:
		getBlocksMsg := new(GetBlocksMessage)
		if err := s.Decoder.Decode(bytes.NewReader(msg.Data), getBlocksMsg); err != nil {
			return nil, fmt.Errorf("failed to decode getblocks message: %w", err)
		}
		decodedMsg.Data = getBlocksMsg

	case MessageTypeBlocks:
		blocksMsg := new(BlocksMessage)
		if err := s.Decoder.Decode(bytes.NewReader(msg.Data), blocksMsg); err != nil {
			return nil, fmt.Errorf("failed to decode blocks message: %w", err)
		}
		decodedMsg.Data = blocksMsg

	default:
		return nil, fmt.Errorf("unknown message header: %v", msg.Header)
	}

	return decodedMsg, nil
}

func (s *Server) SendMessage(to NetAddr, payload []byte) error {
	for _, tr := range s.Transports {
		// 这里假设 Transport 知道如何处理 NetAddr。
		// TCPTransport 会在其 peer map 中查找。
		// 如果一个 Server 连接了多个 transport, 需要 Transport 层能区分 peer。
		// 当前设计是每个 Transport 维护自己的 peer 列表，所以我们尝试通过每个 transport 发送。
		// 如果找到对应的 peer，SendMessage 会成功。
		if err := tr.SendMessage(to, payload); err == nil {
			return nil
		}
	}
	return fmt.Errorf("failed to send message to %s: peer not found in any transport", to)
}
