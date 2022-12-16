package wallet_test

import (
	"encoding/hex"
	"testing"

	"github.com/equitas-foundation/bamp-ocean/pkg/wallet"
	"github.com/stretchr/testify/require"
)

func TestEstimateTxSize(t *testing.T) {
	tests := []struct {
		inputs       []wallet.Input
		outputs      []wallet.Output
		expectedSize int
	}{
		// https://blockstream.info/liquid/tx/3bf5b21f9b5785de089be6dc4963058b4734bf86a9434c9910ad739dbf742eb0
		{
			inputs: []wallet.Input{
				{Script: h2b("16001483a220425cf9f653175e81e7438f57ba7483e262")},
			},
			outputs: []wallet.Output{
				{Script: h2b("a914ecc84d1102e0d11f3ab9448e994248291a3582df87"), BlindingKey: make([]byte, 33)},
				{Script: h2b("a914a24595654e08709c2fd3f68061a313a5821bdef887"), BlindingKey: make([]byte, 33)},
			},
			expectedSize: 2516,
		},
		// https://blockstream.info/liquid/tx/06d4897d60128cccc588ccd5e1d62eba3d23b154ce8954e6b8057356c9eb9fa0
		{
			inputs: []wallet.Input{
				{Script: h2b("1600142fab026d04a9f8534b5ee1c93129a085527ba7b8")},
				{Script: h2b("160014d33bf8fad507164c32bd28ea6e295fa0f4bfee95")},
			},
			outputs: []wallet.Output{
				{Script: h2b("00140a169de9d6f628331f312edb2d9ea0db4fe5233f"), BlindingKey: make([]byte, 33)},
				{Script: h2b("0014b3929f5b7c466b8a6c2d3788e5a25bb67945d3c3"), BlindingKey: make([]byte, 33)},
			},
			expectedSize: 2621,
		},
		// https://blockstream.info/liquid/tx/34941db50a2128008451304200e396b64b68120f411f0a4fe0c2f9cef1f9864f
		{
			inputs: []wallet.Input{
				{Script: h2b("00140b51f5036527f61a234015ed3bdc84497793b26d")},
				{Script: h2b("0014ae71e8f6d614c5df51711760edfd7146e7fb9d74")},
				{Script: h2b("00143020f079534e777b49b8278558405c928951c1fe")},
			},
			outputs: []wallet.Output{
				{Script: h2b("00148fcf009ef09cad277621239c3cacdb57d292030c"), BlindingKey: make([]byte, 33)},
				{Script: h2b("0014824b8f3123053ff69781e686f7f0d0d76416b2ee"), BlindingKey: make([]byte, 33)},
				{Script: h2b("0014016adf9bd0b29edec8fe92698772d42047bcabe3"), BlindingKey: make([]byte, 33)},
				{Script: h2b("0014217cb363fdee0ddfff89c89ece02b07d18671c3b"), BlindingKey: make([]byte, 33)},
				{Script: h2b("0014723819358cefcae8606b1ab398738f86b260b4d1"), BlindingKey: make([]byte, 33)},
			},
			expectedSize: 6258,
		},
		// https://blockstream.info/liquid/tx/14a920f9af73e3f9e34fcb4707b1cccd0adca86e27003a32ed77184d4d41d0f6
		{
			inputs: []wallet.Input{
				{Script: h2b("002026b2fb5626d5dbb0e089fd99ecbbf11561bc2a63b05401d3a27e24c7f2ee9cc5"), ScriptSigSize: 35, WitnessSize: 223},
			},
			outputs: []wallet.Output{
				{Script: h2b("001442b76d36c376d402edb61f239c8d1dc28f3a23f6"), BlindingKey: make([]byte, 33)},
			},
			expectedSize: 1373,
		},
		// example tx with lots of P2WPKH ins and outs
		{
			inputs: []wallet.Input{
				{Script: h2b("00140b51f5036527f61a234015ed3bdc84497793b26d")},
				{Script: h2b("00140b51f5036527f61a234015ed3bdc84497793b26d")},
				{Script: h2b("00140b51f5036527f61a234015ed3bdc84497793b26d")},
				{Script: h2b("00140b51f5036527f61a234015ed3bdc84497793b26d")},
				{Script: h2b("00140b51f5036527f61a234015ed3bdc84497793b26d")},
				{Script: h2b("00140b51f5036527f61a234015ed3bdc84497793b26d")},
				{Script: h2b("00140b51f5036527f61a234015ed3bdc84497793b26d")},
			},
			outputs: []wallet.Output{
				{Script: h2b("00148fcf009ef09cad277621239c3cacdb57d292030c"), BlindingKey: make([]byte, 33)},
				{Script: h2b("00148fcf009ef09cad277621239c3cacdb57d292030c"), BlindingKey: make([]byte, 33)},
				{Script: h2b("00148fcf009ef09cad277621239c3cacdb57d292030c"), BlindingKey: make([]byte, 33)},
				{Script: h2b("00148fcf009ef09cad277621239c3cacdb57d292030c"), BlindingKey: make([]byte, 33)},
				{Script: h2b("00148fcf009ef09cad277621239c3cacdb57d292030c"), BlindingKey: make([]byte, 33)},
			},
			expectedSize: 6532,
		},
	}
	for _, tt := range tests {
		size := wallet.EstimateTxSize(tt.inputs, tt.outputs)
		require.GreaterOrEqual(t, int(size), tt.expectedSize)
	}
}

func h2b(str string) []byte {
	buf, _ := hex.DecodeString(str)
	return buf
}
