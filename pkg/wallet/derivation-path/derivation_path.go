package path

import (
	"fmt"
	"math"
	"math/big"
	"strings"

	"github.com/btcsuite/btcd/btcutil/hdkeychain"
)

// DerivationPath is the data structure representing an HD path.
type DerivationPath []uint32

// ParseDerivationPath converts a derivation path in string format to a
// DerivationPath type.
func ParseDerivationPath(strPath string) (DerivationPath, error) {
	return parseDerivationPath(strPath, false)
}

func ParseRootDerivationPath(strPath string) (DerivationPath, error) {
	path, err := parseDerivationPath(strPath, true)
	if err != nil {
		return nil, err
	}
	if len(path) != 2 {
		return nil, ErrInvalidRootPathLen
	}
	if path[0] < hdkeychain.HardenedKeyStart || path[1] < hdkeychain.HardenedKeyStart {
		return nil, ErrInvalidRootPath
	}
	return path, nil
}

func (path DerivationPath) String() string {
	if len(path) <= 0 {
		return ""
	}

	result := "m"
	for _, component := range path {
		var hardened bool
		if component >= hdkeychain.HardenedKeyStart {
			component -= hdkeychain.HardenedKeyStart
			hardened = true
		}
		result = fmt.Sprintf("%s/%d", result, component)
		if hardened {
			result += "'"
		}
	}
	return result
}

func parseDerivationPath(
	strPath string, checkAbsolutePath bool,
) (DerivationPath, error) {
	if strPath == "" {
		return nil, ErrMissingDerivationPath
	}

	elems := strings.Split(strPath, "/")
	if containsEmptyString(elems) {
		return nil, ErrMalformedDerivationPath
	}
	if checkAbsolutePath {
		if elems[0] != "m" {
			return nil, ErrRequiredAbsoluteDerivationPath
		}
	}
	if len(elems) < 2 {
		return nil, ErrMalformedDerivationPath
	}
	if strings.TrimSpace(elems[0]) == "m" {
		elems = elems[1:]
	}

	path := make(DerivationPath, 0)
	for _, elem := range elems {
		elem = strings.TrimSpace(elem)
		var value uint32

		if strings.HasSuffix(elem, "'") {
			value = hdkeychain.HardenedKeyStart
			elem = strings.TrimSpace(strings.TrimSuffix(elem, "'"))
		}

		// use big int for convertion
		bigval, ok := new(big.Int).SetString(elem, 0)
		if !ok {
			return nil, fmt.Errorf("invalid elem '%s' in path", elem)
		}

		max := math.MaxUint32 - value
		if bigval.Sign() < 0 || bigval.Cmp(big.NewInt(int64(max))) > 0 {
			if value == 0 {
				return nil, fmt.Errorf("elem %v must be in range [0, %d]", bigval, max)
			}
			return nil, fmt.Errorf("elem %v must be in hardened range [0, %d]", bigval, max)
		}
		value += uint32(bigval.Uint64())

		path = append(path, value)
	}

	return path, nil
}

func containsEmptyString(composedPath []string) bool {
	for _, s := range composedPath {
		if s == "" {
			return true
		}
	}
	return false
}
