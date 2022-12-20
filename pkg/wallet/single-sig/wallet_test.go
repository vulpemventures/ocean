package singlesig_test

import (
	"strings"
	"testing"

	wallet "github.com/equitas-foundation/bamp-ocean/pkg/wallet/single-sig"
	"github.com/stretchr/testify/require"
)

const (
	testRootPath = "m/84'/1'"
)

func TestNewWallet(t *testing.T) {
	t.Parallel()

	t.Run("valid", func(t *testing.T) {
		t.Parallel()

		w, err := wallet.NewWallet(wallet.NewWalletArgs{RootPath: testRootPath})
		require.NoError(t, err)

		mnemonic, err := w.Mnemonic()
		require.NoError(t, err)

		otherWallet, err := wallet.NewWalletFromMnemonic(
			wallet.NewWalletFromMnemonicArgs{
				RootPath: testRootPath,
				Mnemonic: mnemonic,
			},
		)
		require.NoError(t, err)

		require.Equal(t, *w, *otherWallet)
	})

	t.Run("invalid", func(t *testing.T) {
		tests := []struct {
			args wallet.NewWalletFromMnemonicArgs
			err  error
		}{
			{
				args: wallet.NewWalletFromMnemonicArgs{
					Mnemonic: strings.Split("legal winner thank year wave sausage worth useful legal winner thank yellow", " "),
				},
				err: wallet.ErrMissingRootPath,
			},
			{
				args: wallet.NewWalletFromMnemonicArgs{
					RootPath: testRootPath,
				},
				err: wallet.ErrMissingMnemonic,
			},
			{
				args: wallet.NewWalletFromMnemonicArgs{
					RootPath: testRootPath,
					Mnemonic: strings.Split("legal winner thank year wave sausage worth useful legal winner thank yellow yellow", " "),
				},
				err: wallet.ErrInvalidMnemonic,
			},
		}
		for _, tt := range tests {
			_, err := wallet.NewWalletFromMnemonic(tt.args)
			require.EqualError(t, tt.err, err.Error())
		}
	})
}
