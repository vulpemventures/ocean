package multisig_test

import (
	"strings"
	"testing"

	wallet "github.com/equitas-foundation/bamp-ocean/pkg/wallet/multi-sig"
	"github.com/stretchr/testify/require"
)

var (
	testRootPath = "m/48'/1'/0'/2'"
	xpubs        = []string{"xpub6EuX7TBEwhFgifQY24vFeMRqeWHGyGCupztDxk7G2ECAqGQ22Fik8E811p8GrM2LfajQzLidXy4qECxhdcxChkjiKhnq2fiVMVjdfSoZQwg"}
)

func TestNewWallet(t *testing.T) {
	t.Parallel()

	t.Run("valid", func(t *testing.T) {
		t.Parallel()

		w, err := wallet.NewWallet(wallet.NewWalletArgs{
			RootPath: testRootPath,
			Xpubs:    xpubs,
		})
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
					Mnemonic: strings.Split("legal winner thank year wave sausage worth useful legal winner thank yellow", " "),
					RootPath: testRootPath,
				},
				err: wallet.ErrMissingCosignerXpubs,
			},
			{
				args: wallet.NewWalletFromMnemonicArgs{
					RootPath: testRootPath,
					Mnemonic: strings.Split("legal winner thank year wave sausage worth useful legal winner thank yellow yellow", " "),
				},
				err: wallet.ErrInvalidMnemonic,
			},
			{
				args: wallet.NewWalletFromMnemonicArgs{
					Mnemonic: strings.Split("legal winner thank year wave sausage worth useful legal winner thank yellow", " "),
					RootPath: testRootPath,
					Xpubs:    []string{"not a valid xpub"},
				},
				err: wallet.ErrInvalidXpub,
			},
		}
		for _, tt := range tests {
			_, err := wallet.NewWalletFromMnemonic(tt.args)
			require.EqualError(t, tt.err, err.Error())
		}
	})
}
