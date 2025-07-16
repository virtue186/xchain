package network

import (
	"fmt"
	"io"
)

// MessageType 定义了消息的类型
const (
	MessageTypeTx    MessageType = 0x1
	MessageTypeBlock MessageType = 0x2
)

type MessageType byte

// Message 是网络上传输的原始消息结构
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

// Peer 是对一个网络对等节点的通用接口
type Peer interface {
	io.Closer          // Peer应该可以被关闭
	Send([]byte) error // Peer应该可以发送数据
}

// HandshakeFunc 负责在节点间建立连接后执行握手协议
type HandshakeFunc func(Peer) error

// NOPHandshakeFunc 是一个空操作的握手函数
func NOPHandshakeFunc(Peer) error { return nil }

// RPC 代表一个从远端节点接收到的原始远程过程调用
type RPC struct {
	From    NetAddr
	Message *Message
}

// DecodedMessage 代表一个已经被解码、包含具体业务数据的消息
type DecodedMessage struct {
	From NetAddr
	Data any // Data可以是*core.Transaction, *core.Block等
}

// RPCProcessor 是处理已解码消息的接口
// 我们的 ChainService 就实现了这个接口
type RPCProcessor interface {
	ProcessMessage(*DecodedMessage) error
}

// NOPRPCProcessor 是一个空操作的处理器，用于测试或默认情况
type NOPRPCProcessor struct{}

func (p *NOPRPCProcessor) ProcessMessage(msg *DecodedMessage) error {
	fmt.Printf("NOP processor received message from %s\n", msg.From)
	return nil
}
