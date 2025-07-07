package network

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

func TestBroadcast(t *testing.T) {
	tra := NewLocalTransport("A")
	trb := NewLocalTransport("B")
	trc := NewLocalTransport("C")

	tra.Connect(trb)
	tra.Connect(trc)

	msg := []byte("Hello World")
	assert.Nil(t, tra.Broadcast(msg))

	rpcb := <-trb.Consume()
	b, err := ioutil.ReadAll(rpcb.Payload)
	assert.Nil(t, err)
	assert.Equal(t, msg, b)

	rpcc := <-trc.Consume()
	c, err := ioutil.ReadAll(rpcc.Payload)
	assert.Nil(t, err)
	assert.Equal(t, msg, c)

}
