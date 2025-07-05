package types

import "encoding/hex"

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
