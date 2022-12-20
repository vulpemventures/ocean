package path_test

import (
	"testing"

	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	path "github.com/equitas-foundation/bamp-ocean/pkg/wallet/derivation-path"
	"github.com/stretchr/testify/require"
)

func TestParseDerivationPath(t *testing.T) {
	t.Parallel()

	t.Run("valid", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			derivationPath string
			expected       path.DerivationPath
		}{
			// Plain absolute derivation paths
			{"m/84'/0'/0'/0", path.DerivationPath{hdkeychain.HardenedKeyStart + 84, hdkeychain.HardenedKeyStart, hdkeychain.HardenedKeyStart, 0}},
			{"m/84'/0'/0'/128", path.DerivationPath{hdkeychain.HardenedKeyStart + 84, hdkeychain.HardenedKeyStart, hdkeychain.HardenedKeyStart, 128}},
			{"m/84'/0'/0'/0'", path.DerivationPath{hdkeychain.HardenedKeyStart + 84, hdkeychain.HardenedKeyStart, hdkeychain.HardenedKeyStart, hdkeychain.HardenedKeyStart}},
			{"m/84'/0'/0'/128'", path.DerivationPath{hdkeychain.HardenedKeyStart + 84, hdkeychain.HardenedKeyStart, hdkeychain.HardenedKeyStart + 0, hdkeychain.HardenedKeyStart + 128}},
			{"m/2147483732/2147483648/2147483648/0", path.DerivationPath{hdkeychain.HardenedKeyStart + 84, hdkeychain.HardenedKeyStart, hdkeychain.HardenedKeyStart, 0}},
			{"m/2147483732/2147483648/2147483648/2147483648", path.DerivationPath{hdkeychain.HardenedKeyStart + 84, hdkeychain.HardenedKeyStart, hdkeychain.HardenedKeyStart, hdkeychain.HardenedKeyStart}},

			// Hexadecimal absolute derivation paths
			{"m/0x54'/0x00'/0x00'/0x00", path.DerivationPath{hdkeychain.HardenedKeyStart + 84, hdkeychain.HardenedKeyStart, hdkeychain.HardenedKeyStart, 0}},
			{"m/0x54'/0x00'/0x00'/0x80", path.DerivationPath{hdkeychain.HardenedKeyStart + 84, hdkeychain.HardenedKeyStart, hdkeychain.HardenedKeyStart, 128}},
			{"m/0x54'/0x00'/0x00'/0x00'", path.DerivationPath{hdkeychain.HardenedKeyStart + 84, hdkeychain.HardenedKeyStart, hdkeychain.HardenedKeyStart, hdkeychain.HardenedKeyStart}},
			{"m/0x54'/0x00'/0x00'/0x80'", path.DerivationPath{hdkeychain.HardenedKeyStart + 84, hdkeychain.HardenedKeyStart, hdkeychain.HardenedKeyStart, hdkeychain.HardenedKeyStart + 128}},
			{"m/0x80000054/0x80000000/0x80000000/0x00", path.DerivationPath{hdkeychain.HardenedKeyStart + 84, hdkeychain.HardenedKeyStart, hdkeychain.HardenedKeyStart, 0}},
			{"m/0x80000054/0x80000000/0x80000000/0x80000000", path.DerivationPath{hdkeychain.HardenedKeyStart + 84, hdkeychain.HardenedKeyStart, hdkeychain.HardenedKeyStart, hdkeychain.HardenedKeyStart}},

			// Weird inputs just to ensure they work
			{"	m  /   84			'\n/\n   00	\n\n\t'   /\n0 ' /\t\t	0", path.DerivationPath{hdkeychain.HardenedKeyStart + 84, hdkeychain.HardenedKeyStart, hdkeychain.HardenedKeyStart, 0}},

			// Relative derivation paths
			{"84'/0'/0/0", path.DerivationPath{hdkeychain.HardenedKeyStart + 84, hdkeychain.HardenedKeyStart, 0, 0}},
			{"0'/0/0", path.DerivationPath{hdkeychain.HardenedKeyStart, 0, 0}},
			{"0/0", path.DerivationPath{0, 0}},
		}
		for _, tt := range tests {
			path, err := path.ParseDerivationPath(tt.derivationPath)
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
			{"", path.ErrMissingDerivationPath},               // Empty relative derivation path
			{"m", path.ErrMalformedDerivationPath},            // Empty absolute derivation path
			{"m/", path.ErrMalformedDerivationPath},           // Missing last derivation component
			{"/84'/0'/0'/0", path.ErrMalformedDerivationPath}, // Absolute path without m prefix, might be user error
			{"m/2147483648'", nil},                            // Overflows 32 bit integer (dynamic values on error, not constant)
			{"m/-1'", nil},                                    // Cannot contain negative number (dynamic values on error, not constant)
			{"0", path.ErrMalformedDerivationPath},            // Bad derivation path
		}

		for _, tt := range tests {
			_, err := path.ParseDerivationPath(tt.derivationPath)
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
			expected path.DerivationPath
		}{
			{"m/84'/0'", path.DerivationPath{hdkeychain.HardenedKeyStart + 84, hdkeychain.HardenedKeyStart}},
			{"m/84'/1'", path.DerivationPath{hdkeychain.HardenedKeyStart + 84, hdkeychain.HardenedKeyStart + 1}},
			{"m/44'/0'", path.DerivationPath{hdkeychain.HardenedKeyStart + 44, hdkeychain.HardenedKeyStart}},
			{"m/44'/1'", path.DerivationPath{hdkeychain.HardenedKeyStart + 44, hdkeychain.HardenedKeyStart + 1}},
		}

		for _, tt := range tests {
			path, err := path.ParseRootDerivationPath(tt.rootPath)
			require.NoError(t, err)
			require.Equal(t, tt.expected, path)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		tests := []struct {
			rootPath    string
			expectedErr error
		}{
			{"", path.ErrMissingDerivationPath},
			{"m/84'", path.ErrInvalidRootPathLen},
			{"m/84'/0'/0'", path.ErrInvalidRootPathLen},
			{"m/84'/0", path.ErrInvalidRootPath},
			{"m/84/0'", path.ErrInvalidRootPath},
			{"84'/0'", path.ErrRequiredAbsoluteDerivationPath},
		}

		for _, tt := range tests {
			_, err := path.ParseRootDerivationPath(tt.rootPath)
			require.EqualError(t, tt.expectedErr, err.Error())
		}
	})
}
