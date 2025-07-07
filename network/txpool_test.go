package network

import (
	"github.com/stretchr/testify/assert"
	"github.com/virtue186/xchain/core"
	"strconv"
	"testing"
	"time"
)

func TestTxPool(t *testing.T) {
	pool := NewTxPool()
	assert.Equal(t, pool.Len(), 0)
	transaction := core.NewTransaction([]byte("test"))
	assert.Nil(t, pool.Add(transaction))
	assert.Equal(t, pool.Len(), 1)
	pool.Flush()
	assert.Equal(t, pool.Len(), 0)
}

func TestTxMapSorter(t *testing.T) {
	pool := NewTxPool()
	txlen := 1000
	for i := 0; i < txlen; i++ {
		transaction := core.NewTransaction([]byte(strconv.FormatInt(int64(i), 10)))
		transaction.SetFirstSeen(time.Now().UnixNano())
		assert.Nil(t, pool.Add(transaction))
	}
	assert.Equal(t, pool.Len(), txlen)

	transactions := pool.Transactions()
	for i := 0; i < len(transactions)-1; i++ {
		assert.True(t, transactions[i].GetFirstSeen() < transactions[i+1].GetFirstSeen())
	}

}
