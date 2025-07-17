package core

import "github.com/virtue186/xchain/types"

// State 管理所有账户的状态，并直接与持久化存储交互
type State struct {
	storage Storage
}

// NewState 创建一个新的 State 实例
func NewState(s Storage) *State {
	return &State{
		storage: s,
	}
}

// Put 将一个账户的状态写入数据库
func (s *State) Put(addr types.Address, state *AccountState) error {
	data, err := state.Encode()
	if err != nil {
		return err
	}
	// 在我们的 LevelDBStorage 中，需要一个通用的 Put 方法
	// 我们暂时假设 Storage 接口可以直接存储任意键值对
	// 这需要对 Storage 接口进行扩展
	// 键名使用一个前缀以避免和区块数据冲突
	return s.storage.Put(accountKey(addr), data)
}

// Get 从数据库中获取一个账户的状态
func (s *State) Get(addr types.Address) (*AccountState, error) {
	data, err := s.storage.Get(accountKey(addr))
	if err != nil {
		// 如果错误是因为键不存在，我们返回一个零值账户，而不是错误
		// 这简化了上层逻辑，因为每个地址都“存在”，只是可能是空的
		if err.Error() == "leveldb: not found" { // 依赖于具体的DB实现，不是很好
			return &AccountState{Address: addr, Balance: 0, Nonce: 0}, nil
		}
		return nil, err
	}

	return DecodeAccountState(data)
}

// Delete 从数据库中删除一个账户的状态（较少使用）
func (s *State) Delete(addr types.Address) error {
	return s.storage.Delete(accountKey(addr))
}

// --- 键名辅助函数 ---

const (
	accountPrefix = "a"
)

func accountKey(addr types.Address) []byte {
	return append([]byte(accountPrefix), addr.ToSlice()...)
}
