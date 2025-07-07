package core

import "fmt"

type Validator interface {
	ValidateBlock(*Block) error
}

type BlockValidator struct {
	bc *BlockChain
}

func NewBlockValidator(bc *BlockChain) *BlockValidator {
	return &BlockValidator{bc: bc}
}

func (v BlockValidator) ValidateBlock(b *Block) error {
	if _, err := v.bc.GetHeader(b.Height); err == nil {
		return fmt.Errorf("block %d already exists with hash (%s)", b.Height, b.Hash(BlockHasher{}))
	}

	if b.Height != v.bc.Height()+1 {
		return fmt.Errorf("block %d does not belong to height %d", b.Height, v.bc.Height()+1)
	}

	preheader, err := v.bc.GetHeader(b.Height - 1)
	if err != nil {
		return err
	}
	hash := BlockHasher{}.Hash(preheader)
	if hash != b.PrevBlockHash {

		return fmt.Errorf("the hash of the previous block (%s) is invalid", b.PrevBlockHash)
	}

	if err := b.Verify(); err != nil {
		return err
	}
	return nil
}
