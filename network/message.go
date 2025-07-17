package network

import (
	"github.com/virtue186/xchain/core"
	"io"
)

// StatusMessage 包含了节点的状态信息，这里主要是区块链的高度
type StatusMessage struct {
	ID            string // 节点的ID，便于追踪
	CurrentHeight uint32
}

type GetStatusMessage struct{}

// GetBlocksMessage 用于向远端节点请求区块
type GetBlocksMessage struct {
	// From 指定了请求的起始区块高度（不包含）
	// 如果是 0, 表示从创世块之后开始请求
	From uint32
	// To 指定了请求的结束区块高度（可选）
	// 如果是 0, 表示请求到对方的最新区块
	To uint32
}

// BlocksMessage 用于响应 GetBlocksMessage，包含了具体的区块数据
type BlocksMessage struct {
	Blocks [][]byte // 序列化后的区块列表
}

// Encode 将 Message 编码到指定的写入器
func (m *Message) Encode(w io.Writer, enc core.Encoder[*Message]) error {
	return enc.Encode(w, m)
}

// Decode 从读取器解码 Message
func (m *Message) Decode(r io.Reader, dec core.Decoder[*Message]) error {
	return dec.Decode(r, m)
}
