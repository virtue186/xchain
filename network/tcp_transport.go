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
	conn     net.Conn // 底层的TCP连接
	outbound bool     // true 表示是我们主动发起的连接；false 表示是对方连接进来的
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		conn:     conn,
		outbound: outbound,
	}
}

// Close 关闭底层的连接。
func (p *TCPPeer) Close() error {
	return p.conn.Close()
}

// TCPTransportOpts 包含了创建 TCPTransport 所需的配置项。
type TCPTransportOpts struct {
	ListenAddr    string                 // 本地监听地址, e.g. ":3000"
	HandshakeFunc HandshakeFunc          // 用于在连接建立后进行握手
	Decoder       core.Decoder[*Message] // 用于从网络流中解码消息的解码器
}

// TCPTransport 实现了 Transport 接口，用于处理TCP网络通信。
type TCPTransport struct {
	TCPTransportOpts
	listener net.Listener
	rpcCh    chan RPC

	lock  sync.RWMutex
	peers map[NetAddr]*TCPPeer
}

// NewTCPTransport 创建一个新的 TCPTransport 实例。
func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: opts,
		rpcCh:            make(chan RPC, 1024), // 创建带缓冲的通道
		peers:            make(map[NetAddr]*TCPPeer),
	}
}

// Consume 返回一个只读的 RPC 通道，上层服务可以从中消费接收到的消息。
// 这是 Transport 接口的实现。
func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcCh
}

// Close 关闭监听器，停止接受新的连接。
// 这是 Transport 接口的实现。
func (t *TCPTransport) Close() error {
	return t.listener.Close()
}

// SendMessage 方法实现，以满足 Transport 接口
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

// Addr 返回本地节点的监听地址。
// 这是 Transport 接口的实现。
func (t *TCPTransport) Addr() NetAddr {
	return NetAddr(t.ListenAddr)
}

// Dial 主动连接到网络中的另一个节点。
// 这是 Transport 接口的实现。
func (t *TCPTransport) Dial(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}

	// 为新建立的连接启动一个独立的 goroutine 进行处理
	go t.handleConn(conn, true)

	return nil
}

// ListenAndAccept 开始在配置的地址上监听并接受入站连接。
func (t *TCPTransport) ListenAndAccept() error {
	var err error
	t.listener, err = net.Listen("tcp", t.ListenAddr)
	if err != nil {
		return err
	}

	// 启动一个独立的 goroutine 来处理接受连接的循环
	go t.startAcceptLoop()

	logrus.Infof("TCP transport listening on port: %s\n", t.ListenAddr)
	return nil
}

// Broadcast 向所有已连接的对等节点广播数据。
// 这是 Transport 接口的实现。
func (t *TCPTransport) Broadcast(payload []byte) error {
	t.lock.RLock()
	defer t.lock.RUnlock()

	for _, peer := range t.peers {
		// 直接通过 peer.conn 发送，效率更高
		if _, err := peer.conn.Write(payload); err != nil {
			// 在实际项目中，这里可能不应立即返回错误，而是记录日志并继续尝试向其他节点广播
			logrus.Errorf("failed to broadcast to peer %s: %v", peer.conn.RemoteAddr(), err)
		}
	}
	return nil
}

// startAcceptLoop 是一个无限循环，用于接受新的入站TCP连接。
func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			// 如果监听器被关闭，会产生一个错误，此时应该退出循环
			logrus.Errorf("TCP accept error: %s\n", err)
			continue
		}
		// 为每一个新接受的连接启动一个独立的 goroutine 进行处理
		go t.handleConn(conn, false)
	}
}

// handleConn 是管理单个TCP连接生命周期的核心函数。
func (t *TCPTransport) handleConn(conn net.Conn, outbound bool) {
	var err error
	peer := NewTCPPeer(conn, outbound)
	addr := NetAddr(conn.RemoteAddr().String())

	// 在函数退出时，确保关闭连接并从peers列表中移除
	defer func() {
		logrus.Infof("dropping peer connection: %s, reason: %v", addr, err)
		t.removePeer(addr)
		conn.Close()
	}()

	// 执行握手协议
	if err = t.HandshakeFunc(peer); err != nil {
		return // 握手失败，直接返回，defer 会执行清理工作
	}

	// 握手成功后，将该节点添加到通讯录中
	t.addPeer(peer)

	// 进入读取循环
	for {
		msg := new(Message)
		// 使用注入的解码器从连接中解码RPC消息（阻塞操作）
		err = t.Decoder.Decode(conn, msg)
		if err != nil {
			// 发生任何错误（如连接关闭、数据格式错误），则退出循环
			return
		}

		// 3. 将解码后的消息包装成 RPC 发送给上层
		rpc := RPC{
			From:    addr,
			Message: msg,
		}
		t.rpcCh <- rpc
	}
}

// addPeer 是一个线程安全的函数，用于添加一个新的对等节点到通讯录。
func (t *TCPTransport) addPeer(peer *TCPPeer) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.peers[NetAddr(peer.conn.RemoteAddr().String())] = peer
}

// removePeer 是一个线程安全的函数，用于从通讯录中移除一个对等节点。
func (t *TCPTransport) removePeer(addr NetAddr) {
	t.lock.Lock()
	defer t.lock.Unlock()
	delete(t.peers, addr)
}
