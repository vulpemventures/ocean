package wallet

import (
	"fmt"

	"github.com/btcsuite/btcd/btcutil/hdkeychain"
)

var (
	ErrMissingNetwork           = fmt.Errorf("missing network")
	ErrMissingMnemonic          = fmt.Errorf("missing mnemonic")
	ErrMissingSigningMasterKey  = fmt.Errorf("missing signing master key")
	ErrMissingBlindingMasterKey = fmt.Errorf("missing blinding master key")
	ErrMissingDerivationPath    = fmt.Errorf("missing derivation path")
	ErrMissingOutputScript      = fmt.Errorf("missing output script")
	ErrMissingPset              = fmt.Errorf("missing pset base64")
	ErrMissingDerivationPaths   = fmt.Errorf("missing derivation path map")

	ErrInvalidEntropySize             = fmt.Errorf("entropy size must be a multiple of 32 in the range [128,256]")
	ErrInvalidMnemonic                = fmt.Errorf("blinding mnemonic is invalid")
	ErrInvalidRootPathLen             = fmt.Errorf("invalid root path length, must be in the form \"m/purpose'/coin_type'\"")
	ErrInvalidRootPath                = fmt.Errorf("root path must contain only hardended values")
	ErrRequiredAbsoluteDerivationPath = fmt.Errorf("path must be an absolute derivation starting with 'm/'")
	ErrInvalidDerivationPathLength    = fmt.Errorf("derivation path must be a relative path in the form \"account'/branch/index\"")
	ErrInvalidDerivationPathAccount   = fmt.Errorf("derivation path's account (first elem) must be hardened (suffix ')")
	ErrInvalidSignatures              = fmt.Errorf("transaction contains invalid signature(s)")

	ErrMalformedDerivationPath         = fmt.Errorf("path must not start or end with a '/'")
	ErrOutOfRangeDerivationPathAccount = fmt.Errorf("account index must be in hardened range [0', %d']", hdkeychain.HardenedKeyStart-1)

	ErrMissingPrevOuts          = fmt.Errorf("missing prevouts")
	ErrInputMissingTxid         = fmt.Errorf("input is missing txid")
	ErrInputInvalidTxid         = fmt.Errorf("invalid input txid length: must be exactly 32 bytes")
	ErrOutputMissingAsset       = fmt.Errorf("output is missing asset")
	ErrOutputInvalidAsset       = fmt.Errorf("invalid output asset length: must be exactly 32 bytes")
	ErrOutputInvalidScript      = fmt.Errorf("invalid output script")
	ErrOutputInvalidBlindingKey = fmt.Errorf("invalid output blinding key")
	ErrMissingInputs            = fmt.Errorf("at least one input is mandatory to create a partial transaction with one or more confidential outputs")
	ErrMissingOwnedInputs       = fmt.Errorf("missing list of owned inputs")
	ErrBlindInvalidInputIndex   = fmt.Errorf("input index to blind is out of range")
	ErrMissingRootPath          = fmt.Errorf("missing root derivation path")
)
