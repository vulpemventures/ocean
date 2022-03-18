package wallet_test

import (
	"crypto/rand"
	"encoding/hex"
	"math/big"
)

func randomScript() []byte {
	return append([]byte{0, 20}, randomBytes(20)...)
}

func randomValueCommitment() []byte {
	return append([]byte{9}, randomBytes(32)...)
}

func randomAssetCommitment() []byte {
	return append([]byte{10}, randomBytes(32)...)
}

func randomHex(len int) string {
	return hex.EncodeToString(randomBytes(len))
}

func randomVout() uint32 {
	return uint32(randomIntInRange(0, 15))
}

func randomValue() uint64 {
	return uint64(randomIntInRange(1000000, 10000000000))
}

func randomBytes(len int) []byte {
	b := make([]byte, len)
	rand.Read(b)
	return b
}

func randomIntInRange(min, max int) int {
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(max)))
	return int(int(n.Int64())) + min
}
