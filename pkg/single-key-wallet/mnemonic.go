package wallet

type NewMnemonicArgs struct {
	EntropySize uint32
}

func (a NewMnemonicArgs) validate() error {
	if a.EntropySize > 0 {
		if a.EntropySize < 128 || a.EntropySize > 256 || a.EntropySize%32 != 0 {
			return ErrInvalidEntropySize
		}
	}
	return nil
}

// NewMnemonic returns a new mnemonic as a list of words:
// 	* EntropySize: 256 -> 24-words mnemonic.
// 	* EntropySize: 128 -> 12-words mnemonic.
func NewMnemonic(args NewMnemonicArgs) ([]string, error) {
	if err := args.validate(); err != nil {
		return nil, err
	}
	if args.EntropySize == 0 {
		args.EntropySize = 256
	}

	return generateMnemonic(int(args.EntropySize))
}
