package wallet_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	wallet "github.com/vulpemventures/ocean/pkg/single-key-wallet"
)

func TestNewWallet(t *testing.T) {
	t.Parallel()

	t.Run("valid", func(t *testing.T) {
		t.Parallel()

		w, err := wallet.NewWallet(wallet.NewWalletArgs{})
		require.NoError(t, err)

		mnemonic, err := w.Mnemonic()
		require.NoError(t, err)

		otherWallet, err := wallet.NewWalletFromMnemonic(
			wallet.NewWalletFromMnemonicArgs{Mnemonic: mnemonic},
		)
		require.NoError(t, err)

		require.Equal(t, *w, *otherWallet)
	})

	t.Run("invalid", func(t *testing.T) {
		tests := []struct {
			mnemonic []string
			err      error
		}{
			{
				mnemonic: nil,
				err:      wallet.ErrMissingMnemonic,
			},
			{
				mnemonic: strings.Split("legal winner thank year wave sausage worth useful legal winner thank yellow yellow", " "),
				err:      wallet.ErrInvalidMnemonic,
			},
		}
		for _, tt := range tests {
			_, err := wallet.NewWalletFromMnemonic(wallet.NewWalletFromMnemonicArgs{
				Mnemonic: tt.mnemonic,
			})
			require.EqualError(t, tt.err, err.Error())
		}
	})
}
