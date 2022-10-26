package wallet_test

import (
	"crypto/rand"
	"encoding/hex"
	"math/big"

	"github.com/vulpemventures/go-elements/address"
	wallet "github.com/vulpemventures/ocean/pkg/single-key-wallet"
)

func randomInputs(num int) []wallet.Input {
	ins := make([]wallet.Input, 0, num)
	for i := 0; i < num; i++ {
		ins = append(ins, wallet.Input{
			TxID:            randomHex(32),
			TxIndex:         randomVout(),
			Value:           randomValue(),
			Asset:           randomHex(32),
			Script:          randomScript(),
			ValueBlinder:    randomBytes(32),
			AssetBlinder:    randomBytes(32),
			ValueCommitment: randomValueCommitment(),
			AssetCommitment: randomAssetCommitment(),
			Nonce:           randomBytes(33),
		})
	}
	return ins
}

func randomOutputs(num int) []wallet.Output {
	outs := make([]wallet.Output, 0, num)
	for i := 0; i < num; i++ {
		addr, _ := address.FromConfidential(testAddresses[i%3])
		outs = append(outs, wallet.Output{
			Asset:       randomHex(32),
			Amount:      randomValue(),
			Script:      addr.Script,
			BlindingKey: addr.BlindingKey,
		})
	}

	return outs
}

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
