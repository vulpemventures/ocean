package path

import (
	"fmt"
)

var (
	ErrMissingDerivationPath          = fmt.Errorf("missing derivation path")
	ErrInvalidRootPathLen             = fmt.Errorf(`invalid root path length, must be in the form m/purpose'/coin_type'`)
	ErrInvalidRootPath                = fmt.Errorf("root path must contain only hardended values")
	ErrRequiredAbsoluteDerivationPath = fmt.Errorf("path must be an absolute derivation starting with 'm/'")
	ErrMalformedDerivationPath        = fmt.Errorf("path must not start or end with a '/'")
)
