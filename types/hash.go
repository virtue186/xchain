package types

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

type Hash [32]uint8

func (h Hash) IsZero() bool {
	for i := 0; i < 32; i++ {
		if h[i] != 0 {
			return false
		}
	}
	return true
}

func (h Hash) ToSlice() []byte {
	b := make([]byte, 32)
	copy(b, h[:])
	return b
}

func (h Hash) String() string {
	return hex.EncodeToString(h.ToSlice())
}

func HashFromBytes(b []byte) Hash {
	if len(b) != 32 {
		panic("hash length must be 32 bytes")
	}
	var h Hash
	copy(h[:], b)
	return Hash(h)
}

func RandomBytes(size int) []byte {
	token := make([]byte, size)
	_, err := rand.Read(token)
	if err != nil {
		return nil
	}
	return token
}

func RandomHash() Hash {
	return HashFromBytes(RandomBytes(32))
}

// MarshalJSON 实现了 json.Marshaler 接口
// 它将 Hash 类型转换为一个带 "0x" 前缀的十六进制字符串
func (h Hash) MarshalJSON() ([]byte, error) {
	// 使用 fmt.Sprintf 创建带双引号的JSON字符串
	return []byte(fmt.Sprintf("\"%s\"", h.String())), nil
}

// UnmarshalJSON 实现了 json.Unmarshaler 接口
// 它可以将一个JSON字符串（带或不带 "0x" 前缀）解析为 Hash 类型
func (h *Hash) UnmarshalJSON(data []byte) error {
	// 1. 去掉字符串首尾的双引号
	if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
		return fmt.Errorf("invalid hash json string: %s", data)
	}
	s := string(data[1 : len(data)-1])

	// 2. 如果有 "0x" 前缀，则去掉 (在 String() 方法中我们不加，但解析时最好兼容)
	if len(s) > 2 && s[:2] == "0x" {
		s = s[2:]
	}

	// 3. 将十六进制字符串解码为字节
	b, err := hex.DecodeString(s)
	if err != nil {
		return err
	}
	if len(b) != 32 {
		return fmt.Errorf("invalid hash length, expected 32 bytes, got %d", len(b))
	}

	// 4. 将解码后的字节复制到 Hash 中
	copy(h[:], b)
	return nil
}
