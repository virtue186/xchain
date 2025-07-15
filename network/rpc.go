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

func DefaultRPCDecodeFunc(dec core.Decoder[any], r io.Reader) (*DecodedMessage, error) {
	// 1. 解码外层的 Message 对象
	msg := &Message{}
	// 使用注入的解码器来解码
	if err := dec.Decode(r, msg); err != nil {
		return nil, fmt.Errorf("decode message error: %w", err)
	}

	// 2. 根据消息头，解码内层的 Data
	switch msg.Header {
	case MessageTypeTx:
		tx := new(core.Transaction)
		// 同样使用注入的解码器来解码 msg.Data
		// 我们需要将 []byte 包装成 io.Reader
		if err := dec.Decode(bytes.NewReader(msg.Data), tx); err != nil {
			return nil, fmt.Errorf("decode transaction error: %w", err)
		}
		return &DecodedMessage{
			Data: tx,
		}, nil

	case MessageTypeBlock:
		b := new(core.Block)
		// 同样使用注入的解码器
		if err := dec.Decode(bytes.NewReader(msg.Data), b); err != nil {
			return nil, fmt.Errorf("decode block error: %w", err)
		}
		return &DecodedMessage{
			Data: b,
		}, nil

	default:
		return nil, fmt.Errorf("unknown message header: %v", msg.Header)
	}
}

type RPCProcessor interface {
	ProcessMessage(*DecodedMessage) error
}
