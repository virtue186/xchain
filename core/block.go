package core

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"github.com/virtue186/xchain/types"
	"io"
)

type Header struct {
	Version   uint32
	PrevHash  types.Hash
	Timestamp int64
	Height    uint32
	Nonce     uint64
}
type Block struct {
	Header
	Transactions []Transaction
	// cached version of the header hash
	hash types.Hash
}

func (b *Block) Hash() types.Hash {
	buf := &bytes.Buffer{}
	b.Header.EncodeBinary(buf)
	b.hash = types.Hash(sha256.Sum256(buf.Bytes()))
	return b.hash
}

func (header *Header) EncodeBinary(w io.Writer) error {

	if err := binary.Write(w, binary.LittleEndian, &header.Version); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, &header.PrevHash); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, &header.Timestamp); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, &header.Height); err != nil {
		return err
	}
	return binary.Write(w, binary.LittleEndian, &header.Nonce)

}

func (header *Header) DecodeBinary(r io.Reader) error {
	if err := binary.Read(r, binary.LittleEndian, &header.Version); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &header.PrevHash); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &header.Timestamp); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &header.Height); err != nil {
		return err
	}
	return binary.Read(r, binary.LittleEndian, &header.Nonce)
}

func (block *Block) EncodeBinary(w io.Writer) error {
	if err := block.Header.EncodeBinary(w); err != nil {
		return err
	}
	for _, tx := range block.Transactions {
		err := tx.EncodeBinary(w)
		if err != nil {
			return err
		}
	}
	return nil
}

func (block *Block) DecodeBinary(r io.Reader) error {
	if err := block.Header.DecodeBinary(r); err != nil {
		return err
	}
	for _, tx := range block.Transactions {
		err := tx.DecodeBinary(r)
		if err != nil {
			return err
		}
	}
	return nil
}
