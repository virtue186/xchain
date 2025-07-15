package network

//
//import (
//	"fmt"
//	"github.com/sirupsen/logrus"
//	"github.com/virtue186/xchain/core"
//	"net"
//	"sync"
//)
//
//type TCPPeer struct {
//	conn     net.Conn
//	outbound bool //outbound: true 表示是我们主动拨出去的连接；false 表示是远端拨入的
//}
//
//func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
//	return &TCPPeer{conn: conn, outbound: outbound}
//}
//
//func (p *TCPPeer) Close() error {
//	return p.conn.Close()
//}
//
//type TCPTransportOpts struct {
//	ListenAddr string // 本地监听地址
//	//HandshakeFunc HandshakeFunc     // 握手函数
//	Decoder core.Decoder[RPC] // 解码函数，用于从流中解出 RPC 消息
//}
//type TCPTransport struct {
//	TCPTransportOpts
//	listener  net.Listener
//	consumeCh chan RPC
//	lock      sync.RWMutex
//
//	peers map[NetAddr]*TCPPeer
//}
//
//func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
//	return &TCPTransport{
//		TCPTransportOpts: opts,
//		peers:            make(map[NetAddr]*TCPPeer),
//		consumeCh:        make(chan RPC, 1024),
//	}
//}
//
//func (t *TCPTransport) Consume() <-chan RPC {
//	return t.consumeCh
//}
//func (t *TCPTransport) Close() error {
//	return t.listener.Close()
//}
//
//func (t *TCPTransport) Addr() NetAddr {
//	return NetAddr(t.ListenAddr)
//}
//
//func (t *TCPTransport) Dial(addr net.Addr) error {
//	conn, err := net.Dial("tcp", addr.String())
//	if err != nil {
//		return err
//	}
//	go t.handleConn(conn, true)
//	return nil
//}
//
//func (t *TCPTransport) Broadcast(data []byte) error {
//	t.lock.RLock()
//	defer t.lock.RUnlock()
//
//	for _, peer := range t.peers {
//		err := t.SendMessage(NetAddr(peer.conn.RemoteAddr().String()), data)
//		if err != nil {
//			return err
//		}
//	}
//
//	return nil
//}
//
//func (t *TCPTransport) SendMessage(to NetAddr, payload []byte) error {
//	t.lock.RLock()
//	defer t.lock.RUnlock()
//	peer, ok := t.peers[to]
//	if !ok {
//		return fmt.Errorf("%s: could not find peer", t.ListenAddr)
//	}
//	_, err := peer.conn.Write(payload)
//	return err
//}
//
//func (t *TCPTransport) ListenAndAccept() error {
//	var err error
//	t.listener, err = net.Listen("tcp", t.ListenAddr)
//	if err != nil {
//		return err
//	}
//	go t.startAcceptLoop()
//	logrus.Infof("TCP transport listening on port: %s\n", t.ListenAddr)
//	return nil
//}
//
//func (t *TCPTransport) startAcceptLoop() {
//	for {
//		conn, err := t.listener.Accept()
//		if err != nil {
//			logrus.Errorf("TCP accept error: %s\n", err)
//			continue
//		}
//		go t.handleConn(conn, false)
//	}
//}
//
//func (t *TCPTransport) handleConn(conn net.Conn, outbound bool) {
//	var err error
//
//	defer func() {
//		fmt.Printf("dropping peer connection: %s", err)
//		conn.Close()
//	}()
//
//	peer := NewTCPPeer(conn, outbound)
//
//	//if err = t.HandshakeFunc(peer); err != nil {
//	//	return
//	//}
//
//	// Read loop
//	rpc := RPC{}
//	for {
//		err = t.Decoder.Decode(conn, &rpc)
//		if err != nil {
//			return
//		}
//
//		rpc.From = NetAddr(conn.RemoteAddr().String())
//		t.consumeCh <- rpc
//	}
//}
