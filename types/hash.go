package types

import (
	"crypto/rand"
	"encoding/hex"
)

type Hash [32]uint8

func (h Hash) IsZero() bool {
	for i := 0; i < 32; i++ {
		if h[i] == 0 {
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
