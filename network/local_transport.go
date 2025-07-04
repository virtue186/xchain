package network

import (
	"fmt"
	"sync"
)

type LocalTransport struct {
	addr      NetAddr
	consumeCh chan RPC
	lock      sync.RWMutex
	peers     map[NetAddr]*LocalTransport
}

func NewLocalTransport(addr NetAddr) Transport {
	return &LocalTransport{
		addr:      addr,
		consumeCh: make(chan RPC, 1024),
		peers:     make(map[NetAddr]*LocalTransport),
	}

}

func (t *LocalTransport) Consume() <-chan RPC {
	return t.consumeCh
}
func (t *LocalTransport) Connect(transport Transport) error {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.peers[transport.Addr()] = transport.(*LocalTransport)
	return nil
}

func (t *LocalTransport) Addr() NetAddr {
	return t.addr
}

func (t *LocalTransport) SendMessage(to NetAddr, payload []byte) error {
	t.lock.RLock()
	defer t.lock.RUnlock()

	trans, ok := t.peers[to]
	if !ok {
		return fmt.Errorf("no transport for %s", to)
	}
	trans.consumeCh <- RPC{
		From:    t.addr,
		Payload: payload,
	}
	return nil
}
