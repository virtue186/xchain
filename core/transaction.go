package core

import (
	"encoding/json"
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
	// 创建一个不包含签名和公钥的临时副本用于签名
	// 这是为了防止签名数据包含上一次的签名结果
	txCopy := *tx
	txCopy.Signature = nil
	txCopy.From = nil

	return json.Marshal(&txCopy)
}

// Encode 将整个交易编码到写入器
func (tx *Transaction) Encode(w io.Writer, enc Encoder[*Transaction]) error {
	return enc.Encode(w, tx)
}

// Decode 从读取器解码交易
func (tx *Transaction) Decode(r io.Reader, dec Decoder[*Transaction]) error {
	return dec.Decode(r, tx)
}
