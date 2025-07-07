package core

type Storage interface {
	Put(*Block) error
}

type MemoryStorage struct {
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{}
}

func (s MemoryStorage) Put(block *Block) error {
	return nil
}
