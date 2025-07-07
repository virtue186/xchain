package main

import (
	"bytes"
	"github.com/sirupsen/logrus"
	"github.com/virtue186/xchain/core"
	"github.com/virtue186/xchain/crypto"
	"github.com/virtue186/xchain/network"
	"math/rand"
	"strconv"
	"time"
)

func main() {

	transport := network.NewLocalTransport("LOCAL")
	transport2 := network.NewLocalTransport("Remote")

	transport.Connect(transport2)
	transport2.Connect(transport)

	go func() {
		for {
			err := sendTransaction(transport2, transport.Addr())
			if err != nil {
				logrus.Error(err)
			}
			time.Sleep(1 * time.Second)
		}
	}()

	opts := network.ServerOpts{
		Transports: []network.Transport{transport},
	}
	server := network.NewServer(opts)
	server.Start()
}

func sendTransaction(tr network.Transport, addr network.NetAddr) error {
	privateKey := crypto.GeneratePrivateKey()
	data := []byte(strconv.FormatInt(int64(rand.Intn(1000)), 10))
	tx := core.NewTransaction(data)
	tx.Sign(privateKey)
	buf := &bytes.Buffer{}
	if err := tx.Encode(core.NewGobTxEncoder(buf)); err != nil {
		return err
	}

	msg := network.NewMessage(network.MessageTypeTx, buf.Bytes())

	return tr.SendMessage(addr, msg.Bytes())
}
