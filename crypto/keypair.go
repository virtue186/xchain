package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"github.com/virtue186/xchain/types"
	"math/big"
)

type PublicKey struct {
	Key *ecdsa.PublicKey
}

type PrivateKey struct {
	key *ecdsa.PrivateKey
}

type Signature struct {
	S *big.Int
	R *big.Int
}

func (k PrivateKey) Sign(data []byte) (*Signature, error) {
	sign, b, err := ecdsa.Sign(rand.Reader, k.key, data)
	if err != nil {
		return nil, err
	}
	return &Signature{sign, b}, nil
}

func (sign Signature) Verify(key PublicKey, data []byte) bool {
	verify := ecdsa.Verify(key.Key, data, sign.S, sign.R)
	return verify
}

func GeneratePrivateKey() PrivateKey {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(err)
	}
	return PrivateKey{key: privateKey}
}

func (k PrivateKey) PublicKey() PublicKey {
	return PublicKey{
		Key: &k.key.PublicKey,
	}
}

func (k PublicKey) ToSlice() []byte {
	return elliptic.MarshalCompressed(k.Key, k.Key.X, k.Key.Y)
}

func (k PublicKey) Address() types.Address {
	h := sha256.Sum256(k.ToSlice())

	return types.AddressFromBytes(h[len(h)-20:])

}
