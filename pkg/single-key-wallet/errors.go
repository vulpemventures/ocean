package wallet

import (
	"errors"
	"fmt"

	"github.com/btcsuite/btcd/btcutil/hdkeychain"
)

var (
	ErrMissingNetwork           = errors.New("missing network")
	ErrMissingMnemonic          = errors.New("missing mnemonic")
	ErrMissingSigningMasterKey  = errors.New("missing signing master key")
	ErrMissingBlindingMasterKey = errors.New("missing blinding master key")
	ErrMissingDerivationPath    = errors.New("missing derivation path")
	ErrMissingOutputScript      = errors.New("missing output script")
	ErrMissingPset              = errors.New("missing pset base64")
	ErrMissingDerivationPaths   = errors.New("missing derivation path map")

	ErrInvalidEntropySize = errors.New(
		"entropy size must be a multiple of 32 in the range [128,256]",
	)
	ErrInvalidMnemonic    = errors.New("blinding mnemonic is invalid")
	ErrInvalidRootPathLen = errors.New(
		"invald root path length. Must be in the form \"m/coin_type'/purpose'\"",
	)
	ErrInvalidRootPathValue = errors.New(
		"root path must contain only hardended values",
	)
	ErrInvalidDerivationPath       = errors.New("invalid derivation path")
	ErrInvalidDerivationPathLength = errors.New(
		"derivation path must be a relative path in the form \"account'/branch/index\"",
	)
	ErrInvalidDerivationPathAccount = errors.New(
		"derivation path's account (first elem) must be hardened (suffix \"'\")",
	)
	ErrInvalidSignatures = errors.New("transaction contains invalid signature(s)")

	ErrMalformedDerivationPath = errors.New(
		"path must not start or end with a '/' and " +
			"can optionally start with 'm/' for absolute paths",
	)
	ErrOutOfRangeDerivationPathAccount = fmt.Errorf(
		"account index must be in hardened range [0', %d']",
		hdkeychain.HardenedKeyStart-1,
	)

	ErrMissingPrevOuts        = errors.New("missing prevouts")
	ErrInputMissingTxid       = errors.New("input is missing txid")
	ErrInputInvalidTxid       = errors.New("invalid input txid length: must be exactly 32 bytes")
	ErrOutputMissingAsset     = errors.New("output is missing asset")
	ErrOutputInvalidAsset     = errors.New("invalid output asset length: must be exactly 32 bytes")
	ErrMissingInputs          = errors.New("at least one input is mandatory to create a partial transaction with one or more confidential outputs")
	ErrMissingOwnedInputs     = errors.New("missing list of owned inputs")
	ErrBlindInvalidInputIndex = errors.New("input index to blind is out of range")
)
