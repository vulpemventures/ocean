package wallet_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vulpemventures/go-elements/psetv2"
	wallet "github.com/vulpemventures/ocean/pkg/single-key-wallet"
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

		w, err := wallet.NewWallet(wallet.NewWalletArgs{})
		require.NoError(t, err)
		require.NotNil(t, w)

		psetBase64, err := w.CreatePset(wallet.CreatePsetArgs{
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
		outs = append(outs, wallet.Output{
			Asset:   randomHex(32),
			Amount:  randomValue(),
			Address: testAddresses[i%3],
		})
	}

	return outs
}
