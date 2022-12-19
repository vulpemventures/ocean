package wallet_test

import (
	"crypto/rand"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vulpemventures/go-elements/address"
	"github.com/vulpemventures/go-elements/psetv2"
	wallet "github.com/vulpemventures/ocean/pkg/wallet"
)

var (
	testAddresses = []string{
		"el1qqfttsemg4sapwrfmmccyztj4wa8gpn5yfetkda4z5uy5e2jysgrszmj0xa8tzftde78kvtl26dtxw6q6gcuawte5xeyvkunws",
		"AzpjXSNnwaFpQQwf2A8AUj6Axqa3YXokJtEwmNvQWvoGn2ymKUzmofHmjxBKzPr7bszjrEJRpPSgJqUp",
		"CTExJqr9PvAveGHmK3ymA3YVdBFvEWh1Vqkj5U9DCv4L46BJhhAd3g8SdjPNCZR268VnsaynRGmyzrQa",
	}
)

func TestCreatePset(t *testing.T) {
	t.Parallel()

	t.Run("valid", func(t *testing.T) {
		inputs := randomInputs(2)
		outputs := randomOutputs(3)

		psetBase64, err := wallet.CreatePset(wallet.CreatePsetArgs{
			Inputs:  inputs,
			Outputs: outputs,
		})
		require.NoError(t, err)
		require.NotEmpty(t, psetBase64)

		ptx, err := psetv2.NewPsetFromBase64(psetBase64)
		require.NoError(t, err)
		require.NotNil(t, ptx)
	})
}

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
