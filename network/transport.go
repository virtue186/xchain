package network

type NetAddr string

type Transport interface {
	Dial(string) error
	Consume() <-chan RPC
	Close() error
	SendMessage(NetAddr, []byte) error
	Broadcast([]byte) error
	Addr() NetAddr
}
