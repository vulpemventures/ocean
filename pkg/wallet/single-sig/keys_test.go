package singlesig_test

import (
	"encoding/hex"
	"testing"

	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/stretchr/testify/require"
	"github.com/vulpemventures/go-elements/network"
	wallet "github.com/vulpemventures/ocean/pkg/wallet/single-sig"
)

func TestAccountExtendedKey(t *testing.T) {
	t.Parallel()

	t.Run("valid", func(t *testing.T) {
		t.Parallel()

		w, err := wallet.NewWallet(wallet.NewWalletArgs{RootPath: testRootPath})
		require.NoError(t, err)

		args := wallet.ExtendedKeyArgs{
			Account: 0,
		}
		xprv, err := w.AccountExtendedPrivateKey(args)
		require.NoError(t, err)
		require.NotEmpty(t, xprv)

		xpub, err := w.AccountExtendedPublicKey(args)
		require.NoError(t, err)
		require.NotEmpty(t, xpub)
	})

	t.Run("invalid", func(t *testing.T) {
		t.Parallel()

		w, err := wallet.NewWallet(wallet.NewWalletArgs{RootPath: testRootPath})
		require.NoError(t, err)

		tests := []struct {
			args wallet.ExtendedKeyArgs
			err  error
		}{
			{
				args: wallet.ExtendedKeyArgs{
					Account: hdkeychain.HardenedKeyStart,
				},
				err: wallet.ErrOutOfRangeDerivationPathAccount,
			},
		}

		for _, tt := range tests {
			_, err := w.AccountExtendedPrivateKey(tt.args)
			require.EqualError(t, tt.err, err.Error())
			_, err = w.AccountExtendedPublicKey(tt.args)
			require.EqualError(t, tt.err, err.Error())
		}
	})
}

func TestDeriveSigningKeyPair(t *testing.T) {
	t.Parallel()

	t.Run("valid", func(t *testing.T) {
		t.Parallel()

		w, err := wallet.NewWallet(wallet.NewWalletArgs{RootPath: testRootPath})
		require.NoError(t, err)

		args := wallet.DeriveSigningKeyPairArgs{
			DerivationPath: "0'/0/0",
		}
		prvkey, pubkey, err := w.DeriveSigningKeyPair(args)
		require.NoError(t, err)
		require.NotNil(t, prvkey)
		require.NotNil(t, pubkey)
	})

	t.Run("invalid", func(t *testing.T) {
		t.Parallel()

		w, err := wallet.NewWallet(wallet.NewWalletArgs{RootPath: testRootPath})
		require.NoError(t, err)

		tests := []struct {
			args wallet.DeriveSigningKeyPairArgs
			err  error
		}{
			{
				args: wallet.DeriveSigningKeyPairArgs{"0/0"},
				err:  wallet.ErrInvalidDerivationPathLength,
			},
			{
				args: wallet.DeriveSigningKeyPairArgs{"0/0/0/0"},
				err:  wallet.ErrInvalidDerivationPathLength,
			},
			{
				args: wallet.DeriveSigningKeyPairArgs{"0'/0/0/0"},
				err:  wallet.ErrInvalidDerivationPathLength,
			},
			{
				args: wallet.DeriveSigningKeyPairArgs{"0/0/0"},
				err:  wallet.ErrInvalidDerivationPathAccount,
			},
		}

		for _, tt := range tests {
			_, _, err := w.DeriveSigningKeyPair(tt.args)
			require.EqualError(t, tt.err, err.Error())
		}
	})
}

func TestDeriveBlindingKeyPair(t *testing.T) {
	t.Parallel()

	t.Run("valid", func(t *testing.T) {
		t.Parallel()

		w, err := wallet.NewWallet(wallet.NewWalletArgs{RootPath: testRootPath})
		require.NoError(t, err)

		script, _ := hex.DecodeString("001439397080b51ef22c59bd7469afacffbeec0da12e")
		args := wallet.DeriveBlindingKeyPairArgs{
			Script: script,
		}
		blindingPrvkey, blindingPubkey, err := w.DeriveBlindingKeyPair(args)
		require.NoError(t, err)
		require.NotNil(t, blindingPrvkey)
		require.NotNil(t, blindingPubkey)
	})

	t.Run("invalid", func(t *testing.T) {
		t.Parallel()

		w, err := wallet.NewWallet(wallet.NewWalletArgs{RootPath: testRootPath})
		if err != nil {
			t.Fatal(err)
		}

		tests := []struct {
			args wallet.DeriveBlindingKeyPairArgs
			err  error
		}{
			{
				args: wallet.DeriveBlindingKeyPairArgs{[]byte{}},
				err:  wallet.ErrMissingOutputScript,
			},
		}

		for _, tt := range tests {
			_, _, err := w.DeriveBlindingKeyPair(tt.args)
			require.EqualError(t, tt.err, err.Error())
		}
	})
}

func TestDeriveConfidentialAddress(t *testing.T) {
	t.Parallel()

	t.Run("valid", func(t *testing.T) {
		t.Parallel()

		w, err := wallet.NewWallet(wallet.NewWalletArgs{RootPath: testRootPath})
		require.NoError(t, err)

		args := wallet.DeriveConfidentialAddressArgs{
			DerivationPath: "0'/0/0",
			Network:        &network.Liquid,
		}
		ctAddress, script, err := w.DeriveConfidentialAddress(args)
		require.NoError(t, err)
		require.NotNil(t, ctAddress)
		require.NotNil(t, script)
	})

	t.Run("invalid", func(t *testing.T) {
		t.Parallel()

		w, err := wallet.NewWallet(wallet.NewWalletArgs{RootPath: testRootPath})
		require.NoError(t, err)

		tests := []struct {
			args wallet.DeriveConfidentialAddressArgs
			err  error
		}{
			{
				args: wallet.DeriveConfidentialAddressArgs{
					DerivationPath: "",
					Network:        &network.Liquid,
				},
				err: wallet.ErrMissingDerivationPath,
			},
			{
				args: wallet.DeriveConfidentialAddressArgs{
					DerivationPath: "0'/0/0",
					Network:        nil,
				},
				err: wallet.ErrMissingNetwork,
			},
		}

		for _, tt := range tests {
			_, _, err := w.DeriveConfidentialAddress(tt.args)
			require.EqualError(t, tt.err, err.Error())
		}
	})
}
