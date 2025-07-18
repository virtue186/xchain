package network

//
//import (
//	"bytes"
//	"fmt"
//	"sync"
//)
//
//type LocalTransport struct {
//	addr      NetAddr
//	consumeCh chan RPC
//	lock      sync.RWMutex
//	peers     map[NetAddr]*LocalTransport
//}
//
//func NewLocalTransport(addr NetAddr) Transport {
//	return &LocalTransport{
//		addr:      addr,
//		consumeCh: make(chan RPC, 1024),
//		peers:     make(map[NetAddr]*LocalTransport),
//	}
//
//}
//
//func (t *LocalTransport) Consume() <-chan RPC {
//	return t.consumeCh
//}
//func (t *LocalTransport) Connect(transport Transport) error {
//	t.lock.Lock()
//	defer t.lock.Unlock()
//
//	t.peers[transport.Addr()] = transport.(*LocalTransport)
//	return nil
//}
//
//func (t *LocalTransport) Addr() NetAddr {
//	return t.addr
//}
//
//func (t *LocalTransport) SendMessage(to NetAddr, payload []byte) error {
//	t.lock.RLock()
//	defer t.lock.RUnlock()
//
//	trans, ok := t.peers[to]
//	if !ok {
//		return fmt.Errorf("no peer send for %s", to)
//	}
//	trans.consumeCh <- RPC{
//		From:    t.addr,
//		Payload: bytes.NewReader(payload),
//	}
//	return nil
//}
//
//func (t *LocalTransport) Broadcast(payload []byte) error {
//	for _, peer := range t.peers {
//		if err := t.SendMessage(peer.addr, payload); err != nil {
//			return err
//		}
//	}
//	return nil
//}
