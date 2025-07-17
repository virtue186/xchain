package core

import (
	"bytes"
	"encoding/gob"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/virtue186/xchain/types"
	"strconv"
)

type LeveldbStorage struct {
	db *leveldb.DB
}

func (s *LeveldbStorage) Put(key, value []byte) error {
	return s.db.Put(key, value, nil)
}

func (s *LeveldbStorage) Get(key []byte) ([]byte, error) {
	return s.db.Get(key, nil)
}

func (s *LeveldbStorage) Delete(key []byte) error {
	return s.db.Delete(key, nil)
}

func NewLeveldbStorage(path string) (*LeveldbStorage, error) {
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return nil, err
	}
	return &LeveldbStorage{db: db}, nil
}

func (s *LeveldbStorage) PutBlock(block *Block) error {
	buf := new(bytes.Buffer)
	err := gob.NewEncoder(buf).Encode(block)
	if err != nil {
		return err
	}
	blockHash := block.Hash(BlockHasher{})
	batch := new(leveldb.Batch)

	batch.Put(blockHeightKey(block.Height), blockHash.ToSlice())
	batch.Put(blockKey(blockHash), buf.Bytes())
	return s.db.Write(batch, nil)

}

// GetBlockByHash 根据区块哈希从数据库中获取区块
func (s *LeveldbStorage) GetBlockByHash(hash types.Hash) (*Block, error) {
	data, err := s.db.Get(blockKey(hash), nil)
	if err != nil {
		return nil, err
	}

	return DecodeBlock(data)
}

// GetBlockHashByHeight 根据区块高度从数据库中获取区块哈希
func (s *LeveldbStorage) GetBlockHashByHeight(height uint32) (types.Hash, error) {
	data, err := s.db.Get(blockHeightKey(height), nil)
	if err != nil {
		return types.Hash{}, err
	}

	return types.HashFromBytes(data), nil
}

// Close 关闭数据库连接
func (s *LeveldbStorage) Close() error {
	return s.db.Close()
}

const (
	blockHeightPrefix = "h"
	blockPrefix       = "b"
)

var (
	blockHeightPrefixB = []byte(blockHeightPrefix) // []byte{'h'}
	blockPrefixB       = []byte(blockPrefix)       // []byte{'b'}
)

func blockHeightKey(height uint32) []byte {
	// 字节前缀 + 最多 10 位 uint32 十进制数
	b := make([]byte, 0, 1+10)
	b = append(b, blockHeightPrefixB...)
	b = strconv.AppendUint(b, uint64(height), 10)
	return b
}

func blockKey(hash types.Hash) []byte {
	// 区块哈希通常固定 32字节 ⇒ 预分配 1+32
	b := make([]byte, 0, 1+len(hash))
	b = append(b, blockPrefixB...)
	b = append(b, hash.ToSlice()...)
	return b
}
