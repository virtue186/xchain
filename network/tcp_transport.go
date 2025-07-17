package network

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/virtue186/xchain/core"
	"net"
	"sync"
)

// TCPPeer 代表一个通过 TCP 连接的远端节点。
type TCPPeer struct {
	conn     net.Conn
	outbound bool
}

func (p *TCPPeer) Send(payload []byte) error {
	_, err := p.conn.Write(payload)
	return err
}

// RemoteAddr 实现了 Peer 接口，返回远端节点的地址
func (p *TCPPeer) RemoteAddr() NetAddr {
	return NetAddr(p.conn.RemoteAddr().String())
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		conn:     conn,
		outbound: outbound,
	}
}

func (p *TCPPeer) Close() error {
	return p.conn.Close()
}

// TCPTransportOpts 包含了创建 TCPTransport 所需的配置项。
type TCPTransportOpts struct {
	ListenAddr    string
	HandshakeFunc HandshakeFunc
	Decoder       core.Decoder[*Message]
}

// TCPTransport 实现了 Transport 接口，用于处理TCP网络通信。
type TCPTransport struct {
	TCPTransportOpts
	listener net.Listener
	rpcCh    chan RPC
	peerCh   chan Peer // 用于广播新连接的 Peer

	lock  sync.RWMutex
	peers map[NetAddr]*TCPPeer
}

func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: opts,
		rpcCh:            make(chan RPC, 1024),
		peerCh:           make(chan Peer, 10), // 确保通道被初始化
		peers:            make(map[NetAddr]*TCPPeer),
	}
}

// PeerEvents 实现了 Transport 接口，返回 peer 事件通道
func (t *TCPTransport) PeerEvents() <-chan Peer {
	return t.peerCh
}

func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcCh
}

func (t *TCPTransport) Close() error {
	return t.listener.Close()
}

func (t *TCPTransport) SendMessage(to NetAddr, payload []byte) error {
	t.lock.RLock()
	defer t.lock.RUnlock()

	peer, ok := t.peers[to]
	if !ok {
		return fmt.Errorf("%s: could not find peer %s", t.ListenAddr, to)
	}

	_, err := peer.conn.Write(payload)
	return err
}

func (t *TCPTransport) Addr() NetAddr {
	return NetAddr(t.ListenAddr)
}

func (t *TCPTransport) Dial(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}

	go t.handleConn(conn, true)
	return nil
}

func (t *TCPTransport) ListenAndAccept() error {
	var err error
	t.listener, err = net.Listen("tcp", t.ListenAddr)
	if err != nil {
		return err
	}

	go t.startAcceptLoop()

	logrus.Infof("TCP transport listening on port: %s\n", t.ListenAddr)
	return nil
}

func (t *TCPTransport) Broadcast(payload []byte) error {
	t.lock.RLock()
	defer t.lock.RUnlock()

	for _, peer := range t.peers {
		if _, err := peer.conn.Write(payload); err != nil {
			logrus.Errorf("failed to broadcast to peer %s: %v", peer.conn.RemoteAddr(), err)
		}
	}
	return nil
}

func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			logrus.Errorf("TCP accept error: %s\n", err)
			continue
		}
		go t.handleConn(conn, false)
	}
}

func (t *TCPTransport) handleConn(conn net.Conn, outbound bool) {
	var err error
	peer := NewTCPPeer(conn, outbound)
	addr := NetAddr(conn.RemoteAddr().String())

	defer func() {
		logrus.Infof("dropping peer connection: %s, reason: %v", addr, err)
		t.removePeer(addr)
		conn.Close()
	}()

	if err = t.HandshakeFunc(peer); err != nil {
		return
	}

	// 握手成功后，将该节点添加到通讯录并发出事件
	t.addPeer(peer)

	for {
		msg := new(Message)
		err = t.Decoder.Decode(conn, msg)
		if err != nil {
			return
		}

		rpc := RPC{
			From:    addr,
			Message: msg,
		}
		t.rpcCh <- rpc
	}
}

// addPeer 是一个线程安全的函数，用于添加一个新的对等节点到通讯录
// 这是关键的修改点！
func (t *TCPTransport) addPeer(peer *TCPPeer) {
	t.lock.Lock()
	defer t.lock.Unlock()

	addr := NetAddr(peer.conn.RemoteAddr().String())
	t.peers[addr] = peer

	// 将新建立的 Peer 发送到事件通道
	// 使用非阻塞发送，以防通道已满导致死锁
	select {
	case t.peerCh <- peer:
		logrus.Infof("new peer (%s) sent to peer channel", addr)
	default:
		logrus.Warnf("peer channel is full, dropping peer event for %s", addr)
	}
}

// removePeer 是一个线程安全的函数，用于从通讯录中移除一个对等节点
func (t *TCPTransport) removePeer(addr NetAddr) {
	t.lock.Lock()
	defer t.lock.Unlock()
	delete(t.peers, addr)
}
