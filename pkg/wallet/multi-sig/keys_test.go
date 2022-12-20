package multisig_test

import (
	"encoding/hex"
	"testing"

	wallet "github.com/equitas-foundation/bamp-ocean/pkg/wallet/multi-sig"
	"github.com/stretchr/testify/require"
	"github.com/vulpemventures/go-elements/network"
)

func TestAccountExtendedKey(t *testing.T) {
	t.Parallel()

	t.Run("valid", func(t *testing.T) {
		t.Parallel()

		w, err := wallet.NewWallet(wallet.NewWalletArgs{
			RootPath: testRootPath,
			Xpubs:    xpubs,
		})
		require.NoError(t, err)

		xprv, err := w.AccountExtendedPrivateKey()
		require.NoError(t, err)
		require.NotEmpty(t, xprv)

		xpub, err := w.AccountExtendedPublicKey()
		require.NoError(t, err)
		require.NotEmpty(t, xpub)
	})
}

func TestDeriveSigningKeyPair(t *testing.T) {
	t.Parallel()

	t.Run("valid", func(t *testing.T) {
		t.Parallel()

		w, err := wallet.NewWallet(wallet.NewWalletArgs{
			RootPath: testRootPath,
			Xpubs:    xpubs,
		})
		require.NoError(t, err)

		args := wallet.DeriveSigningKeyPairArgs{
			DerivationPath: "0/0",
		}
		prvkey, pubkeys, err := w.DeriveSigningKeyPair(args)
		require.NoError(t, err)
		require.NotNil(t, prvkey)
		require.NotEmpty(t, pubkeys)
	})

	t.Run("invalid", func(t *testing.T) {
		t.Parallel()

		w, err := wallet.NewWallet(wallet.NewWalletArgs{
			RootPath: testRootPath,
			Xpubs:    xpubs,
		})
		require.NoError(t, err)

		tests := []struct {
			args wallet.DeriveSigningKeyPairArgs
			err  error
		}{
			{
				args: wallet.DeriveSigningKeyPairArgs{"0'/0"},
				err:  wallet.ErrInvalidDerivationPathAccount,
			},
			{
				args: wallet.DeriveSigningKeyPairArgs{"0'/0/0"},
				err:  wallet.ErrInvalidDerivationPathLength,
			},
			{
				args: wallet.DeriveSigningKeyPairArgs{"0/0/0"},
				err:  wallet.ErrInvalidDerivationPathLength,
			},
		}

		for _, tt := range tests {
			privateKey, pubkeys, err := w.DeriveSigningKeyPair(tt.args)
			require.EqualError(t, err, tt.err.Error())
			require.Nil(t, privateKey)
			require.Empty(t, pubkeys)
		}
	})
}

func TestDeriveBlindingKeyPair(t *testing.T) {
	t.Parallel()

	t.Run("valid", func(t *testing.T) {
		t.Parallel()

		w, err := wallet.NewWallet(wallet.NewWalletArgs{
			RootPath: testRootPath,
			Xpubs:    xpubs,
		})
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

		w, err := wallet.NewWallet(wallet.NewWalletArgs{
			RootPath: testRootPath,
			Xpubs:    xpubs,
		})
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
			prvkey, pubkey, err := w.DeriveBlindingKeyPair(tt.args)
			require.EqualError(t, tt.err, err.Error())
			require.Nil(t, prvkey)
			require.Nil(t, pubkey)
		}
	})
}

func TestDeriveConfidentialAddress(t *testing.T) {
	t.Parallel()

	t.Run("valid", func(t *testing.T) {
		t.Parallel()

		w, err := wallet.NewWallet(wallet.NewWalletArgs{
			RootPath: testRootPath,
			Xpubs:    xpubs,
		})
		require.NoError(t, err)

		args := wallet.DeriveConfidentialAddressArgs{
			DerivationPath: "0/0",
			Network:        &network.Liquid,
		}
		ctAddress, script, redeemScript, err := w.DeriveConfidentialAddress(args)
		require.NoError(t, err)
		require.NotNil(t, ctAddress)
		require.NotNil(t, script)
		require.NotNil(t, redeemScript)
	})

	t.Run("invalid", func(t *testing.T) {
		t.Parallel()

		w, err := wallet.NewWallet(wallet.NewWalletArgs{
			RootPath: testRootPath,
			Xpubs:    xpubs,
		})
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
					DerivationPath: "0/0",
					Network:        nil,
				},
				err: wallet.ErrMissingNetwork,
			},
		}

		for _, tt := range tests {
			addr, script, redeemScript, err := w.DeriveConfidentialAddress(tt.args)
			require.EqualError(t, tt.err, err.Error())
			require.Empty(t, addr)
			require.Empty(t, script)
			require.Empty(t, redeemScript)
		}
	})
}
