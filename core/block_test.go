package core

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/virtue186/xchain/types"
	"testing"
	"time"
)

func TestHeader_EncodeDecode(t *testing.T) {

	h := &Header{
		Version:   1,
		PrevHash:  types.RandomHash(),
		Timestamp: time.Now().UnixNano(),
		Height:    10,
		Nonce:     984567,
	}

	buf := &bytes.Buffer{}
	assert.Nil(t, h.EncodeBinary(buf))

	h2 := &Header{}
	assert.Nil(t, h2.DecodeBinary(buf))

	assert.Equal(t, h, h2)
}

func TestBlock_EncodeDecode(t *testing.T) {
	b := &Block{
		Header: Header{
			Version:   1,
			PrevHash:  types.RandomHash(),
			Timestamp: time.Now().UnixNano(),
			Height:    10,
			Nonce:     984567,
		},
		Transactions: nil,
	}

	buf := &bytes.Buffer{}
	assert.Nil(t, b.EncodeBinary(buf))
	b2 := &Block{}
	assert.Nil(t, b2.DecodeBinary(buf))
	assert.Equal(t, b, b2)

	fmt.Printf("%+v\n", b2)
}

func TestBlock_Hash(t *testing.T) {
	b := &Block{
		Header: Header{
			Version:   1,
			PrevHash:  types.RandomHash(),
			Timestamp: time.Now().UnixNano(),
			Height:    10,
			Nonce:     984567,
		},
		Transactions: []Transaction{},
	}

	hash := b.Hash()
	fmt.Printf("%x\n", hash)
	assert.False(t, hash.IsZero())
}
