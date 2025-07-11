package network

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/virtue186/xchain/core"
	"io"
)

const (
	MessageTypeTx    MessageType = 0x1
	MessageTypeBlock MessageType = 0x2
)

type MessageType byte
type Message struct {
	Header MessageType
	Data   []byte
}

func NewMessage(header MessageType, data []byte) *Message {
	return &Message{
		Header: header,
		Data:   data,
	}
}

func (m *Message) Bytes() []byte {
	buf := &bytes.Buffer{}
	gob.NewEncoder(buf).Encode(m)
	return buf.Bytes()
}

type RPC struct {
	From    NetAddr
	Payload io.Reader
}

type DecodedMessage struct {
	From NetAddr
	Data any
}

type RPCDecodeFunc func(RPC) (*DecodedMessage, error)

func DefaultRPCDecodeFunc(rpc RPC) (*DecodedMessage, error) {
	msg := &Message{}
	if err := gob.NewDecoder(rpc.Payload).Decode(msg); err != nil {
		return nil, fmt.Errorf("decode RPC payload error: %w", err)
	}
	switch msg.Header {
	case MessageTypeTx:
		tx := new(core.Transaction)
		if err := tx.Decode(core.NewGobTxDecoder(bytes.NewReader(msg.Data))); err != nil {
			return nil, err
		}
		return &DecodedMessage{
			rpc.From,
			tx,
		}, nil
	case MessageTypeBlock:
		block := new(core.Block)
		err := block.Decode(core.NewGobBlockDecoder(bytes.NewReader(msg.Data)))
		if err != nil {
			return nil, err
		}
		return &DecodedMessage{
			rpc.From,
			block,
		}, nil
	default:
		return nil, fmt.Errorf("unknown RPC type: %v", msg.Header)
	}
}

// 业务层接口
type RPCProcessor interface {
	ProcessMessage(*DecodedMessage) error
}
