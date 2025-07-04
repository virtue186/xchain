package main

import (
	"github.com/virtue186/xchain/network"
	"time"
)

func main() {

	transport := network.NewLocalTransport("LOCAL")
	transport2 := network.NewLocalTransport("Remote")

	transport.Connect(transport2)
	transport2.Connect(transport)

	go func() {
		for {
			transport2.SendMessage(transport.Addr(), []byte("hello world"))
			time.Sleep(1 * time.Second)
		}
	}()

	opts := network.ServerOpts{
		Transports: []network.Transport{transport},
	}
	server := network.NewServer(opts)
	server.Start()
}
