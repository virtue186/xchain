package core

import (
	"encoding/json"
	"github.com/virtue186/xchain/types"
)

type AccountState struct {
	Address types.Address `json:"address"` // 添加json标签以获得更清晰的输出
	Balance uint64        `json:"balance"` // 账户余额
	Nonce   uint64        `json:"nonce"`   // 交易计数器
}

// Encode 将 AccountState 编码为字节流
func (a *AccountState) Encode() ([]byte, error) {
	// 将 gob.NewEncoder(buf).Encode(a) 替换为 json.Marshal(a)
	return json.Marshal(a)
}

// Decode 从字节流解码 AccountState
func DecodeAccountState(b []byte) (*AccountState, error) {
	as := new(AccountState)
	// 将 gob.NewDecoder(...).Decode(as) 替换为 json.Unmarshal(b, as)
	if err := json.Unmarshal(b, as); err != nil {
		return nil, err
	}
	return as, nil
}
