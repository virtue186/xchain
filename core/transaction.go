package core

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/virtue186/xchain/crypto"
	"github.com/virtue186/xchain/types"
	"io"
)

type Transaction struct {
	Data      []byte
	From      crypto.PublicKey // 发送方地址
	Signature *crypto.Signature
	To        types.Address // 接收方地址
	Value     uint64        // 转移的金额
	Nonce     uint64        // 发送方发出的交易序号，用于防止重放攻击

	hash      types.Hash
	firstSeen int64
}

type TxData struct {
	Data  []byte
	To    types.Address
	Value uint64
	Nonce uint64
}

func NewTransaction(data []byte) *Transaction {
	return &Transaction{
		Data: data,
	}
}

func (tx *Transaction) Hash(hasher Hasher[*Transaction]) types.Hash {
	if tx.hash.IsZero() {
		tx.hash = hasher.Hash(tx)
	}
	return tx.hash
}

func (tx *Transaction) Sign(privateKey crypto.PrivateKey) error {

	sigData, err := tx.encodeForSignature()
	if err != nil {
		return err
	}

	sign, err := privateKey.Sign(sigData)
	if err != nil {
		return err
	}

	tx.From = privateKey.PublicKey()
	tx.Signature = sign
	return nil
}

func (tx *Transaction) Verify() error {
	if tx.Signature == nil {
		return fmt.Errorf("transaction signature is nil")
	}

	data, err := tx.encodeForSignature()
	if err != nil {
		return fmt.Errorf("failed to encode for signature: %w", err)
	}

	if !tx.Signature.Verify(tx.From, data) {
		return fmt.Errorf("transaction signature is invalid")
	}
	return nil
}

func (tx *Transaction) SetFirstSeen(firstSeen int64) {
	tx.firstSeen = firstSeen
}

func (tx *Transaction) GetFirstSeen() int64 {
	return tx.firstSeen
}

// encodeForSignature 是一个内部辅助方法，用于将需要签名的数据编码
func (tx *Transaction) encodeForSignature() ([]byte, error) {
	buf := new(bytes.Buffer)

	// 为了确保签名的绝对一致性，我们创建一个临时的、匿名的结构体
	// 它只包含需要被签名的核心字段。
	dataToSign := &struct {
		Data  []byte
		To    types.Address
		Value uint64
		Nonce uint64
	}{
		To:    tx.To,
		Value: tx.Value,
		Nonce: tx.Nonce,
	}

	if len(tx.Data) == 0 {
		dataToSign.Data = []byte{}
	} else {
		dataToSign.Data = tx.Data
	}

	// 对这个规范化后的结构进行编码
	if err := gob.NewEncoder(buf).Encode(dataToSign); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Encode 将整个交易编码到写入器
func (tx *Transaction) Encode(w io.Writer, enc Encoder[*Transaction]) error {
	return enc.Encode(w, tx)
}

// Decode 从读取器解码交易
func (tx *Transaction) Decode(r io.Reader, dec Decoder[*Transaction]) error {
	return dec.Decode(r, tx)
}
