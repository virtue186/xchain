package network

import (
	"github.com/virtue186/xchain/core"
	"github.com/virtue186/xchain/types"
	"sort"
)

type TxPool struct {
	transactions map[types.Hash]*core.Transaction
}

func NewTxPool() *TxPool {
	return &TxPool{
		transactions: make(map[types.Hash]*core.Transaction),
	}
}

func (p *TxPool) Len() int {
	return len(p.transactions)
}

func (p *TxPool) Add(tx *core.Transaction) error {
	hash := tx.Hash(core.TxHasher{})
	p.transactions[hash] = tx
	return nil
}

func (p *TxPool) Flush() {
	p.transactions = make(map[types.Hash]*core.Transaction)
}

func (p *TxPool) Has(hash types.Hash) bool {
	_, ok := p.transactions[hash]
	return ok
}

type TxMapSorter struct {
	transaction []*core.Transaction
}

func NewTxMapSorter(txMap map[types.Hash]*core.Transaction) *TxMapSorter {
	txx := make([]*core.Transaction, len(txMap))
	i := 0
	for _, tx := range txMap {
		txx[i] = tx
		i++
	}
	s := &TxMapSorter{txx}
	sort.Sort(s)
	return s

}

func (s *TxMapSorter) Len() int {
	return len(s.transaction)
}
func (s *TxMapSorter) Swap(i, j int) {
	s.transaction[i], s.transaction[j] = s.transaction[j], s.transaction[i]
}
func (s *TxMapSorter) Less(i, j int) bool {
	return s.transaction[i].GetFirstSeen() < s.transaction[j].GetFirstSeen()
}

func (p *TxPool) Transactions() []*core.Transaction {
	s := NewTxMapSorter(p.transactions)
	return s.transaction
}
