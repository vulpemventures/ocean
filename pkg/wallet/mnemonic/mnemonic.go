package mnemonic

import (
	"fmt"
	"strings"

	"github.com/tyler-smith/go-bip39"
)

var (
	ErrInvalidEntropySize = fmt.Errorf("entropy size must be 128 or 256")
)

type NewMnemonicArgs struct {
	EntropySize uint32
}

func (a NewMnemonicArgs) validate() error {
	if a.EntropySize > 0 {
		if a.EntropySize != 128 && a.EntropySize != 256 {
			return ErrInvalidEntropySize
		}
	}
	return nil
}

// NewMnemonic returns a new mnemonic as a list of words:
//   - EntropySize: 256 -> 24-words mnemonic.
//   - EntropySize: 128 -> 12-words mnemonic.
func NewMnemonic(args NewMnemonicArgs) ([]string, error) {
	if err := args.validate(); err != nil {
		return nil, err
	}
	if args.EntropySize == 0 {
		args.EntropySize = 256
	}

	entropy, err := bip39.NewEntropy(int(args.EntropySize))
	if err != nil {
		return nil, err
	}
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return nil, err
	}
	return strings.Split(mnemonic, " "), nil
}
