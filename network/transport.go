package network

import "io"

type NetAddr string

// Peer 是对一个网络对等节点的通用接口
type Peer interface {
	io.Closer          // Peer应该可以被关闭
	Send([]byte) error // Peer应该可以发送数据
	RemoteAddr() NetAddr
}

type Transport interface {
	Dial(string) error
	Consume() <-chan RPC
	Close() error
	SendMessage(NetAddr, []byte) error
	Broadcast([]byte) error
	Addr() NetAddr
	PeerEvents() <-chan Peer
}
