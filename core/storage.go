package core

import "github.com/virtue186/xchain/types"

type Storage interface {
	Close() error
	// 通用键值对存储
	Put([]byte, []byte) error
	Get([]byte) ([]byte, error)
	Delete([]byte) error

	// 专用于区块的方法
	PutBlock(*Block) error
	GetBlockByHash(types.Hash) (*Block, error)
	GetBlockHashByHeight(uint32) (types.Hash, error)
}
