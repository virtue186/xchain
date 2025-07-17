package types

import (
	"encoding/hex"
	"fmt"
)

type Address [20]uint8

func (addr Address) ToSlice() []byte {
	bytes := make([]byte, 20)
	for i := 0; i < 20; i++ {
		bytes[i] = addr[i]
	}
	return bytes
}

func (addr Address) String() string {
	return hex.EncodeToString(addr.ToSlice())
}

func AddressFromBytes(b []byte) Address {
	if len(b) != 20 {
		panic("length must be 20 bytes")
	}
	var value Address
	for i := 0; i < 20; i++ {
		value[i] = b[i]
	}

	return value
}

func AddressFromHex(s string) (Address, error) {
	// 1. 如果有 "0x" 前缀，则去掉
	if len(s) > 2 && s[:2] == "0x" {
		s = s[2:]
	}

	// 2. 将十六进制字符串解码为字节
	b, err := hex.DecodeString(s)
	if err != nil {
		return Address{}, err
	}

	// 3. 验证长度是否正确
	if len(b) != 20 {
		return Address{}, fmt.Errorf("invalid address length, expected 20 bytes, got %d", len(b))
	}

	// 4. 使用已有的 AddressFromBytes 进行转换
	return AddressFromBytes(b), nil
}
