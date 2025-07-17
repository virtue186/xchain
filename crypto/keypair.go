package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/virtue186/xchain/types"
	"io"
	"math/big"
)

type PrivateKey struct {
	key *ecdsa.PrivateKey
}

func (k PrivateKey) Sign(data []byte) (*Signature, error) {
	hash := sha256.Sum256(data)
	r, s, err := ecdsa.Sign(rand.Reader, k.key, hash[:])
	if err != nil {
		return nil, err
	}

	return &Signature{
		R: r,
		S: s,
	}, nil
}

func NewPrivateKeyFromReader(r io.Reader) PrivateKey {
	key, err := ecdsa.GenerateKey(elliptic.P256(), r)
	if err != nil {
		panic(err)
	}

	return PrivateKey{
		key: key,
	}
}

func GeneratePrivateKey() PrivateKey {
	return NewPrivateKeyFromReader(rand.Reader)
}

func (k PrivateKey) PublicKey() PublicKey {
	return elliptic.MarshalCompressed(k.key.PublicKey, k.key.PublicKey.X, k.key.PublicKey.Y)
}

type PublicKey []byte

func (k PublicKey) String() string {
	return hex.EncodeToString(k)
}

func (k PublicKey) Address() types.Address {
	h := sha256.Sum256(k)

	return types.AddressFromBytes(h[len(h)-20:])
}

type Signature struct {
	S *big.Int
	R *big.Int
}

func (sig Signature) String() string {
	b := append(sig.S.Bytes(), sig.R.Bytes()...)
	return hex.EncodeToString(b)
}

func (sig Signature) Verify(pubKey PublicKey, data []byte) bool {
	x, y := elliptic.UnmarshalCompressed(elliptic.P256(), pubKey)
	key := &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     x,
		Y:     y,
	}
	// 1. 同样对原始数据进行哈希计算，以确保与签名时的数据一致
	hash := sha256.Sum256(data)

	// 2. 用同样的哈希结果进行验证
	return ecdsa.Verify(key, hash[:], sig.R, sig.S)
}

func NewPrivateKeyFromHex(hexKey string) (PrivateKey, error) {
	// 1. 将十六进制字符串解码为字节
	b, err := hex.DecodeString(hexKey)
	if err != nil {
		return PrivateKey{}, err
	}

	// 2. 验证私钥长度是否符合P-256曲线的要求 (通常是32字节)
	if len(b) != 32 {
		return PrivateKey{}, fmt.Errorf("invalid private key length, expected 32 bytes, got %d", len(b))
	}

	// 3. 将字节转换为 ecdsa.PrivateKey
	priv := new(ecdsa.PrivateKey)
	priv.D = new(big.Int).SetBytes(b)
	priv.PublicKey.Curve = elliptic.P256()
	priv.PublicKey.X, priv.PublicKey.Y = priv.PublicKey.Curve.ScalarBaseMult(b)

	return PrivateKey{
		key: priv,
	}, nil
}

func (k PrivateKey) String() string {
	return hex.EncodeToString(k.key.D.Bytes())
}
