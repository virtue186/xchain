package main

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/virtue186/xchain/core"
	"github.com/virtue186/xchain/crypto"
	"github.com/virtue186/xchain/network"
	"log"
	"time"
)

func main() {

	trLocal := network.NewLocalTransport("LOCAL")
	trRemoteA := network.NewLocalTransport("Remote_A")
	trRemoteB := network.NewLocalTransport("Remote_B")
	trRemoteC := network.NewLocalTransport("Remote_C")

	trLocal.Connect(trRemoteA)
	trRemoteA.Connect(trLocal)
	trRemoteA.Connect(trRemoteB)
	trRemoteB.Connect(trRemoteC)

	go func() {
		for {
			err := sendTransaction(trRemoteA, trLocal.Addr())
			if err != nil {
				logrus.Error(err)
			}
			time.Sleep(2 * time.Second)
		}
	}()
	// 开启一个延迟启动的服务
	//go func() {
	//	time.Sleep(30 * time.Second)
	//	trLater := network.NewLocalTransport("Later")
	//	trLocal.Connect(trLater)
	//	trLater.Connect(trLocal)
	//	server2 := makeServer("Later", trLater, nil)
	//	server2.Start()
	//}()

	//  初始化远程服务

	initRemoteServer([]network.Transport{trRemoteA, trRemoteB, trRemoteC})

	// 开启本地服务
	privateKey := crypto.GeneratePrivateKey()
	server := makeServer("LOCAL", trLocal, &privateKey)
	server.Start()

}

func initRemoteServer(transports []network.Transport) {
	for i := 0; i < len(transports); i++ {
		id := fmt.Sprintf("Remote_%d", i)
		server := makeServer(id, transports[i], nil)
		go server.Start()
	}

}

func makeServer(id string, tr network.Transport, key *crypto.PrivateKey) *network.Server {
	opts := network.ServerOpts{
		PrivateKey: key,
		Transports: []network.Transport{tr},
		ID:         id,
	}
	server, err := network.NewServer(opts)
	if err != nil {
		log.Fatal(err)
	}
	return server
}

func sendTransaction(tr network.Transport, addr network.NetAddr) error {
	privateKey := crypto.GeneratePrivateKey()
	code := []byte{
		byte(core.InstrPushByte), 'f', // push 63
		byte(core.InstrPushByte), 'o', // push 62
		byte(core.InstrPushByte), 'o',
		byte(core.InstrPushInt), 3,
		byte(core.InstrPack),
		byte(core.InstrPushInt), 2,
		byte(core.InstrPushInt), 3,
		byte(core.InstrAdd),
		byte(core.InstrStore),
	}
	tx := core.NewTransaction(code)
	tx.Sign(privateKey)
	buf := &bytes.Buffer{}
	encoder := core.GOBEncoder[any]{}

	if err := encoder.Encode(buf, tx); err != nil {
		return err
	}

	msg := network.NewMessage(network.MessageTypeTx, buf.Bytes())

	return tr.SendMessage(addr, msg.Bytes())
}
