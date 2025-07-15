package main

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/virtue186/xchain/core"
	"github.com/virtue186/xchain/crypto"
	"github.com/virtue186/xchain/network"
	"log"
	"math/rand"
	"time"
)

func main() {
	key := crypto.GeneratePrivateKey()
	validatorNode := makeNode("127.0.0.1:3000", &key)
	remoteNodeA := makeNode("127.0.0.1:4000", nil)
	remoteNodeB := makeNode("127.0.0.1:5000", nil)

	go validatorNode.Server.Start()
	go remoteNodeA.Server.Start()
	go remoteNodeB.Server.Start()

	time.Sleep(500 * time.Millisecond)
	if err := remoteNodeA.Transport.Dial(string(validatorNode.Transport.Addr())); err != nil {
		log.Fatalf("failed to dial validator node: %v", err)
	}
	if err := remoteNodeB.Transport.Dial(string(validatorNode.Transport.Addr())); err != nil {
		log.Fatalf("failed to dial validator node: %v", err)
	}
	// 给予短暂时间让连接建立
	time.Sleep(10 * time.Second) // 新的，更可靠的等待时间

	go func() {
		for {
			// 注意：这里需要知道 node1 的地址才能直接发送
			// 在真实的 P2P 网络中，通常是广播而不是定向发送
			// 这里我们为了演示，直接使用 node1 的地址
			if err := sendTransaction(remoteNodeB.Transport, validatorNode.Transport.Addr()); err != nil {
				logrus.Error(err)
			}
			time.Sleep(time.Millisecond * time.Duration(500+rand.Intn(1000)))
		}
	}()

	// 5. 阻塞主程序，让P2P网络持续运行
	select {}

}

// 定义一个辅助结构体，让 makeNode 的返回更清晰
type Node struct {
	Transport network.Transport
	Server    *network.Server
}

func makeNode(listenAddr string, pk *crypto.PrivateKey) *Node {
	// TCPTransport 的配置项
	opts := network.TCPTransportOpts{
		ListenAddr:    listenAddr,
		HandshakeFunc: network.NOPHandshakeFunc,
		Decoder:       core.GOBDecoder[*network.Message]{},
	}
	tr := network.NewTCPTransport(opts)

	// Server 的配置项
	serverOpts := network.ServerOpts{
		PrivateKey: pk,
		Transports: []network.Transport{tr},
		ID:         fmt.Sprintf("NODE-%s", listenAddr),
	}
	server, err := network.NewServer(serverOpts)
	if err != nil {
		log.Fatal(err)
	}

	// 在 Server 创建成功后，再启动 Transport 的监听
	if err := tr.ListenAndAccept(); err != nil {
		log.Fatal(err)
	}

	return &Node{
		Transport: tr,
		Server:    server,
	}
}

func sendTransaction(tr network.Transport, to network.NetAddr) error {
	privateKey := crypto.GeneratePrivateKey()
	//code := []byte{
	//	byte(core.InstrPushByte), 'f',
	//	byte(core.InstrPushByte), 'o',
	//	byte(core.InstrPushByte), 'o',
	//	byte(core.InstrPushInt), 3,
	//	byte(core.InstrPack),
	//	byte(core.InstrPushInt), 2,
	//	byte(core.InstrPushInt), 3,
	//	byte(core.InstrAdd),
	//	byte(core.InstrStore),
	//}
	txData := []byte(fmt.Sprintf("a unique tx data: %d", time.Now().UnixNano()))
	tx := core.NewTransaction(txData)
	tx.Sign(privateKey)

	buf := &bytes.Buffer{}
	encoder := core.GOBEncoder[any]{}

	// 首先将交易编码
	if err := encoder.Encode(buf, tx); err != nil {
		return err
	}

	// 然后将交易数据包装在 Message 中
	msg := network.NewMessage(network.MessageTypeTx, buf.Bytes())

	// 最后将整个 Message 编码后发送
	finalBuf := &bytes.Buffer{}
	if err := encoder.Encode(finalBuf, msg); err != nil {
		return err
	}

	return tr.SendMessage(to, finalBuf.Bytes())
}
