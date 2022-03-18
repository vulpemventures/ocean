package wallet_test

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
	wallet "github.com/vulpemventures/ocean/pkg/single-key-wallet"
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
				{Address: "H4mWpfhRc6FTpbd8aSFxjiAGSccHDo6oSr"},
				{Address: "GwyYE4fNG2Q9NyfYy8bfivgTeWawvGFJXm"},
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
				{Address: "ex1qpgtfm6wk7c5rx8e39mdjm84qmd872geltlcf26"},
				{Address: "ex1qkwff7kmuge4c5mpdx7ywtgjmkeu5t57r6jpghu"},
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
				{Address: "ex1q3l8sp8hsnjkjwa3pywwretxm2lffyqcvn9qwjg"},
				{Address: "ex1qsf9c7vfrq5lld9upu6r00uxs6ajpdvhw54p95u"},
				{Address: "ex1qq94dlx7sk20daj87jf5cwuk5yprme2lr7397k9"},
				{Address: "ex1qy97txclaacxallufez0vuq4s05vxw8pmv6fnk4"},
				{Address: "ex1qwgupjdvval9wscrtr2eesuu0s6expdx3vd4ru6"},
			},
			expectedSize: 6258,
		},
		// https://blockstream.info/liquid/tx/14a920f9af73e3f9e34fcb4707b1cccd0adca86e27003a32ed77184d4d41d0f6
		{
			inputs: []wallet.Input{
				{Script: h2b("22002026b2fb5626d5dbb0e089fd99ecbbf11561bc2a63b05401d3a27e24c7f2ee9cc5")},
			},
			outputs: []wallet.Output{
				{Address: "ex1qg2mk6dkrwm2q9mdkru3eergac28n5glkdft5fr"},
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
				{Address: "ex1q3l8sp8hsnjkjwa3pywwretxm2lffyqcvn9qwjg"},
				{Address: "ex1q3l8sp8hsnjkjwa3pywwretxm2lffyqcvn9qwjg"},
				{Address: "ex1q3l8sp8hsnjkjwa3pywwretxm2lffyqcvn9qwjg"},
				{Address: "ex1q3l8sp8hsnjkjwa3pywwretxm2lffyqcvn9qwjg"},
				{Address: "ex1q3l8sp8hsnjkjwa3pywwretxm2lffyqcvn9qwjg"},
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
