package core

import (
	"bytes"
	"encoding/gob"
	"github.com/virtue186/xchain/types"
)

type AccountState struct {
	Address types.Address
	Balance uint64 // 账户余额
	Nonce   uint64 // 交易计数器
}

// Encode 将 AccountState 编码为字节流
func (a *AccountState) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(a); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Decode 从字节流解码 AccountState
func DecodeAccountState(b []byte) (*AccountState, error) {
	as := new(AccountState)
	if err := gob.NewDecoder(bytes.NewReader(b)).Decode(as); err != nil {
		return nil, err
	}
	return as, nil
}
