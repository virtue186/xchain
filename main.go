package main

import (
	"bytes"
	"fmt"
	"github.com/go-kit/log"
	"github.com/virtue186/xchain/core"
	"github.com/virtue186/xchain/crypto"
	"github.com/virtue186/xchain/network"
	"github.com/virtue186/xchain/node"
	"github.com/virtue186/xchain/types"
	"os"
	"time"
)

func main() {
	validatorKey := crypto.GeneratePrivateKey()

	validatorTr, validatorNode := makeNode("127.0.0.1:3000", &validatorKey)
	remoteATr, remoteANode := makeNode("127.0.0.1:4000", nil)
	remoteBTr, remoteBNode := makeNode("127.0.0.1:5000", nil)

	go validatorNode.Start()
	go remoteANode.Start()
	go remoteBNode.Start()

	time.Sleep(1 * time.Second)

	fmt.Println("Connecting nodes...")
	if err := remoteATr.Dial(string(validatorTr.Addr())); err != nil {
		fmt.Printf("Error dialing validator from node A: %v\n", err)
	}
	if err := remoteBTr.Dial(string(validatorTr.Addr())); err != nil {
		fmt.Printf("Error dialing validator from node B: %v\n", err)
	}

	go func() {
		for {
			// 每隔2-3秒发送一笔交易
			time.Sleep(time.Duration(2000+time.Now().Nanosecond()%1000) * time.Millisecond)
			if err := sendTransaction(remoteATr, validatorTr.Addr()); err != nil {
				fmt.Printf("Error sending transaction: %v\n", err)
			}
		}
	}()

	select {}
}

func NewGenesisBlock() (*core.Block, error) {
	header := &core.Header{
		Version:   1,
		DataHash:  types.Hash{},
		Height:    0,
		Timestamp: 000000,
	}

	b, _ := core.NewBlock(header, nil)
	return b, nil
}

func makeNode(listenAddr string, pk *crypto.PrivateKey) (network.Transport, *node.Node) {
	logger := log.NewLogfmtLogger(os.Stderr)
	logger = log.With(logger, "node", listenAddr)

	genesis, err := NewGenesisBlock()
	if err != nil {
		panic(err)
	}
	bc, err := core.NewBlockChain(log.With(logger, "module", "blockchain"), genesis)
	if err != nil {
		panic(err)
	}
	txPool := network.NewTxPool(1000)

	opts := network.TCPTransportOpts{
		ListenAddr:    listenAddr,
		HandshakeFunc: network.NOPHandshakeFunc,
		Decoder:       core.GOBDecoder[*network.Message]{},
	}
	tr := network.NewTCPTransport(opts)

	go func() {
		if err := tr.ListenAndAccept(); err != nil {
			fmt.Printf("TCP transport on %s failed: %v\n", listenAddr, err)
		}
	}()

	nodeOpts := node.NodeOpts{
		Logger:     logger,
		Transport:  tr,
		BlockChain: bc,
		TxPool:     txPool,
		PrivateKey: pk,
		BlockTime:  5 * time.Second,
	}

	nodeInstance, err := node.NewNode(nodeOpts)
	if err != nil {
		panic(err)
	}

	return tr, nodeInstance
}

func sendTransaction(tr network.Transport, to network.NetAddr) error {

	privateKey := crypto.GeneratePrivateKey()

	txData := []byte(fmt.Sprintf("a unique tx data: %d", time.Now().UnixNano()))
	tx := core.NewTransaction(txData)
	if err := tx.Sign(privateKey); err != nil {
		return err
	}

	buf := &bytes.Buffer{}
	encoder := core.GOBEncoder[any]{}
	if err := encoder.Encode(buf, tx); err != nil {
		return err
	}

	msg := network.NewMessage(network.MessageTypeTx, buf.Bytes())

	finalBuf := &bytes.Buffer{}
	if err := encoder.Encode(finalBuf, msg); err != nil {
		return err
	}

	fmt.Printf("=> Sending new transaction to %s\n", to)
	return tr.SendMessage(to, finalBuf.Bytes())
}
