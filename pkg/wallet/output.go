package wallet

import (
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/vulpemventures/go-elements/address"
	"github.com/vulpemventures/go-elements/elementsutil"
	"github.com/vulpemventures/go-elements/psetv2"
)

var (
	ErrOutputMissingAsset       = fmt.Errorf("output is missing asset")
	ErrOutputInvalidAsset       = fmt.Errorf("invalid output asset length: must be exactly 32 bytes")
	ErrOutputInvalidScript      = fmt.Errorf("invalid output script")
	ErrOutputInvalidBlindingKey = fmt.Errorf("invalid output blinding key")
)

// Output is the data structure representing an output to be added to a partial
// transaction, therefore inclusing asset, amount and address.
type Output psetv2.OutputArgs

func (o Output) Validate() error {
	if o.Asset == "" {
		return ErrOutputMissingAsset
	}
	asset, err := elementsutil.AssetHashToBytes(o.Asset)
	if err != nil {
		return err
	}
	if len(asset) != 33 {
		return ErrOutputInvalidAsset
	}
	if len(o.Script) > 0 {
		if _, err := address.ParseScript(o.Script); err != nil {
			return ErrOutputInvalidScript
		}
	}
	if len(o.BlindingKey) > 0 {
		if _, err := btcec.ParsePubKey(o.BlindingKey); err != nil {
			return ErrOutputInvalidBlindingKey
		}
	}
	return nil
}

func (o Output) IsConfidential() bool {
	return len(o.BlindingKey) > 0
}

func (o Output) ScriptSize() int {
	return varSliceSerializeSize(o.Script)
}
