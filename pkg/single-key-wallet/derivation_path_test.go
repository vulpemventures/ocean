package wallet_test

import (
	"testing"

	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/stretchr/testify/require"
	wallet "github.com/vulpemventures/ocean/pkg/single-key-wallet"
)

func TestParseDerivationPath(t *testing.T) {
	t.Parallel()

	t.Run("valid", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			derivationPath string
			expected       wallet.DerivationPath
		}{
			// Plain absolute derivation paths
			{"m/84'/0'/0'/0", wallet.DerivationPath{hdkeychain.HardenedKeyStart + 84, hdkeychain.HardenedKeyStart, hdkeychain.HardenedKeyStart, 0}},
			{"m/84'/0'/0'/128", wallet.DerivationPath{hdkeychain.HardenedKeyStart + 84, hdkeychain.HardenedKeyStart, hdkeychain.HardenedKeyStart, 128}},
			{"m/84'/0'/0'/0'", wallet.DerivationPath{hdkeychain.HardenedKeyStart + 84, hdkeychain.HardenedKeyStart, hdkeychain.HardenedKeyStart, hdkeychain.HardenedKeyStart}},
			{"m/84'/0'/0'/128'", wallet.DerivationPath{hdkeychain.HardenedKeyStart + 84, hdkeychain.HardenedKeyStart, hdkeychain.HardenedKeyStart + 0, hdkeychain.HardenedKeyStart + 128}},
			{"m/2147483732/2147483648/2147483648/0", wallet.DerivationPath{hdkeychain.HardenedKeyStart + 84, hdkeychain.HardenedKeyStart, hdkeychain.HardenedKeyStart, 0}},
			{"m/2147483732/2147483648/2147483648/2147483648", wallet.DerivationPath{hdkeychain.HardenedKeyStart + 84, hdkeychain.HardenedKeyStart, hdkeychain.HardenedKeyStart, hdkeychain.HardenedKeyStart}},

			// Hexadecimal absolute derivation paths
			{"m/0x54'/0x00'/0x00'/0x00", wallet.DerivationPath{hdkeychain.HardenedKeyStart + 84, hdkeychain.HardenedKeyStart, hdkeychain.HardenedKeyStart, 0}},
			{"m/0x54'/0x00'/0x00'/0x80", wallet.DerivationPath{hdkeychain.HardenedKeyStart + 84, hdkeychain.HardenedKeyStart, hdkeychain.HardenedKeyStart, 128}},
			{"m/0x54'/0x00'/0x00'/0x00'", wallet.DerivationPath{hdkeychain.HardenedKeyStart + 84, hdkeychain.HardenedKeyStart, hdkeychain.HardenedKeyStart, hdkeychain.HardenedKeyStart}},
			{"m/0x54'/0x00'/0x00'/0x80'", wallet.DerivationPath{hdkeychain.HardenedKeyStart + 84, hdkeychain.HardenedKeyStart, hdkeychain.HardenedKeyStart, hdkeychain.HardenedKeyStart + 128}},
			{"m/0x80000054/0x80000000/0x80000000/0x00", wallet.DerivationPath{hdkeychain.HardenedKeyStart + 84, hdkeychain.HardenedKeyStart, hdkeychain.HardenedKeyStart, 0}},
			{"m/0x80000054/0x80000000/0x80000000/0x80000000", wallet.DerivationPath{hdkeychain.HardenedKeyStart + 84, hdkeychain.HardenedKeyStart, hdkeychain.HardenedKeyStart, hdkeychain.HardenedKeyStart}},

			// Weird inputs just to ensure they work
			{"	m  /   84			'\n/\n   00	\n\n\t'   /\n0 ' /\t\t	0", wallet.DerivationPath{hdkeychain.HardenedKeyStart + 84, hdkeychain.HardenedKeyStart, hdkeychain.HardenedKeyStart, 0}},

			// Relative derivation paths
			{"84'/0'/0/0", wallet.DerivationPath{hdkeychain.HardenedKeyStart + 84, hdkeychain.HardenedKeyStart, 0, 0}},
			{"0'/0/0", wallet.DerivationPath{hdkeychain.HardenedKeyStart, 0, 0}},
			{"0/0", wallet.DerivationPath{0, 0}},
		}
		for _, tt := range tests {
			path, err := wallet.ParseDerivationPath(tt.derivationPath)
			require.NoError(t, err)
			require.Equal(t, tt.expected, path)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			derivationPath string
			expectedErr    error
		}{
			// Invalid derivation paths
			{"", wallet.ErrMissingDerivationPath},               // Empty relative derivation path
			{"m", wallet.ErrMalformedDerivationPath},            // Empty absolute derivation path
			{"m/", wallet.ErrMalformedDerivationPath},           // Missing last derivation component
			{"/84'/0'/0'/0", wallet.ErrMalformedDerivationPath}, // Absolute path without m prefix, might be user error
			{"m/2147483648'", nil},                              // Overflows 32 bit integer (dynamic values on error, not constant)
			{"m/-1'", nil},                                      // Cannot contain negative number (dynamic values on error, not constant)
			{"0", wallet.ErrMalformedDerivationPath},            // Bad derivation path
		}

		for _, tt := range tests {
			_, err := wallet.ParseDerivationPath(tt.derivationPath)
			require.Error(t, err)
			if tt.expectedErr != nil {
				require.EqualError(t, tt.expectedErr, err.Error())
			}
		}
	})
}

func TestParseRootDerivationPath(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		tests := []struct {
			rootPath string
			expected wallet.DerivationPath
		}{
			{"m/84'/0'", wallet.DerivationPath{hdkeychain.HardenedKeyStart + 84, hdkeychain.HardenedKeyStart}},
			{"m/84'/1'", wallet.DerivationPath{hdkeychain.HardenedKeyStart + 84, hdkeychain.HardenedKeyStart + 1}},
			{"m/44'/0'", wallet.DerivationPath{hdkeychain.HardenedKeyStart + 44, hdkeychain.HardenedKeyStart}},
			{"m/44'/1'", wallet.DerivationPath{hdkeychain.HardenedKeyStart + 44, hdkeychain.HardenedKeyStart + 1}},
		}

		for _, tt := range tests {
			path, err := wallet.ParseRootDerivationPath(tt.rootPath)
			require.NoError(t, err)
			require.Equal(t, tt.expected, path)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		tests := []struct {
			rootPath    string
			expectedErr error
		}{
			{"", wallet.ErrMissingDerivationPath},
			{"m/84'", wallet.ErrInvalidRootPathLen},
			{"m/84'/0'/0'", wallet.ErrInvalidRootPathLen},
			{"m/84'/0", wallet.ErrInvalidRootPath},
			{"m/84/0'", wallet.ErrInvalidRootPath},
			{"84'/0'", wallet.ErrRequiredAbsoluteDerivationPath},
		}

		for _, tt := range tests {
			_, err := wallet.ParseRootDerivationPath(tt.rootPath)
			require.EqualError(t, tt.expectedErr, err.Error())
		}
	})
}
