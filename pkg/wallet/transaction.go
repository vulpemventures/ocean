package wallet

import (
	"fmt"

	"github.com/vulpemventures/go-elements/confidential"
	"github.com/vulpemventures/go-elements/elementsutil"
	"github.com/vulpemventures/go-elements/psetv2"
)

var (
	ErrMissingPset       = fmt.Errorf("missing pset base64")
	ErrMissingInputs     = fmt.Errorf("at least one input is mandatory to create a partial transaction with one or more confidential outputs")
	ErrInvalidSignatures = fmt.Errorf("transaction contains invalid signature(s)")

	DummyFeeAmount = uint64(700)
)

type CreatePsetArgs struct {
	Inputs  []Input
	Outputs []Output
}

func (a CreatePsetArgs) validate() error {
	for i, in := range a.Inputs {
		if err := in.Validate(); err != nil {
			return fmt.Errorf("invalid input %d: %s", i, err)
		}
	}

	for i, out := range a.Outputs {
		if err := out.Validate(); err != nil {
			return fmt.Errorf("invalid output %d: %s", i, err)
		}
		if out.IsConfidential() && len(a.Inputs) == 0 {
			return ErrMissingInputs
		}
	}

	return nil
}

func (a CreatePsetArgs) inputs() []psetv2.InputArgs {
	ins := make([]psetv2.InputArgs, 0, len(a.Inputs))
	for _, in := range a.Inputs {
		ins = append(ins, psetv2.InputArgs{
			Txid:    in.TxID,
			TxIndex: in.TxIndex,
		})
	}
	return ins
}

func (a CreatePsetArgs) outputs() []psetv2.OutputArgs {
	outs := make([]psetv2.OutputArgs, 0, len(a.Outputs))
	for _, out := range a.Outputs {
		outs = append(outs, psetv2.OutputArgs(out))
	}
	return outs
}

// CreatePset creates a new partial transaction with given inputs and outputs.
func CreatePset(args CreatePsetArgs) (string, error) {
	if err := args.validate(); err != nil {
		return "", err
	}

	ptx, err := psetv2.New(args.inputs(), args.outputs(), nil)
	if err != nil {
		return "", err
	}

	updater, err := psetv2.NewUpdater(ptx)
	if err != nil {
		return "", err
	}
	for i, in := range args.Inputs {
		prevout := in.Prevout()
		updater.AddInWitnessUtxo(i, prevout)
		if prevout.IsConfidential() {
			asset, _ := elementsutil.TxIDToBytes(in.Asset)
			assetProof, _ := confidential.CreateBlindAssetProof(
				asset, in.AssetCommitment, in.AssetBlinder,
			)
			valueProof, _ := confidential.CreateBlindValueProof(
				nil, in.ValueBlinder, in.Value, in.ValueCommitment, in.AssetCommitment,
			)

			updater.AddInUtxoRangeProof(i, in.RangeProof)
			updater.AddInExplicitAsset(i, asset, assetProof)
			updater.AddInExplicitValue(i, in.Value, valueProof)
		}
		if len(in.RedeemScript) > 0 {
			updater.AddInWitnessScript(i, in.RedeemScript)
		}
	}

	return ptx.ToBase64()
}

type UpdatePsetArgs struct {
	PsetBase64 string
	Inputs     []Input
	Outputs    []Output
}

func (a UpdatePsetArgs) validate() error {
	if len(a.PsetBase64) == 0 {
		return ErrMissingPset
	}
	if _, err := psetv2.NewPsetFromBase64(a.PsetBase64); err != nil {
		return err
	}

	for i, in := range a.Inputs {
		if err := in.Validate(); err != nil {
			return fmt.Errorf("invalid input %d: %s", i, err)
		}
	}

	for i, out := range a.Outputs {
		if err := out.Validate(); err != nil {
			return fmt.Errorf("invalid output %d: %s", i, err)
		}
	}
	return nil
}

func (a UpdatePsetArgs) inputs() []psetv2.InputArgs {
	ins := make([]psetv2.InputArgs, 0, len(a.Inputs))
	for _, in := range a.Inputs {
		ins = append(ins, psetv2.InputArgs{
			Txid:    in.TxID,
			TxIndex: in.TxIndex,
		})
	}
	return ins
}

func (a UpdatePsetArgs) outputs(blinderIndex uint32) []psetv2.OutputArgs {
	outs := make([]psetv2.OutputArgs, 0, len(a.Outputs))
	for _, out := range a.Outputs {
		out.BlinderIndex = blinderIndex
		outs = append(outs, psetv2.OutputArgs(out))
	}
	return outs
}

// UpdatesPset adds inputs and outputs to the given partial transaction.
func UpdatePset(args UpdatePsetArgs) (string, error) {
	if err := args.validate(); err != nil {
		return "", err
	}

	ptx, _ := psetv2.NewPsetFromBase64(args.PsetBase64)
	updater, err := psetv2.NewUpdater(ptx)
	if err != nil {
		return "", err
	}

	nextInputIndex := uint32(ptx.Global.InputCount)
	if err := updater.AddInputs(args.inputs()); err != nil {
		return "", err
	}

	for i, in := range args.Inputs {
		prevout := in.Prevout()
		inIndex := int(nextInputIndex) + i
		if err := updater.AddInWitnessUtxo(inIndex, prevout); err != nil {
			return "", err
		}
		if prevout.IsConfidential() {
			asset, _ := elementsutil.TxIDToBytes(in.Asset)
			assetProof, _ := confidential.CreateBlindAssetProof(
				asset, in.AssetCommitment, in.AssetBlinder,
			)
			valueProof, _ := confidential.CreateBlindValueProof(
				nil, in.ValueBlinder, in.Value, in.ValueCommitment, in.AssetCommitment,
			)

			updater.AddInUtxoRangeProof(inIndex, in.RangeProof)
			updater.AddInExplicitAsset(inIndex, asset, assetProof)
			updater.AddInExplicitValue(inIndex, in.Value, valueProof)
		}
		if len(in.RedeemScript) > 0 {
			if err := updater.AddInWitnessScript(inIndex, in.RedeemScript); err != nil {
				return "", err
			}
		}
	}

	if err := updater.AddOutputs(args.outputs(nextInputIndex)); err != nil {
		return "", err
	}

	return ptx.ToBase64()
}

type FinalizeAndExtractTransactionArgs struct {
	PsetBase64 string
}

func (a FinalizeAndExtractTransactionArgs) validate() error {
	if _, err := psetv2.NewPsetFromBase64(a.PsetBase64); err != nil {
		return err
	}
	return nil
}

// FinalizeAndExtractTransaction attempts to finalize the provided partial
// transaction and eventually extracts the final raw transaction, returning
// it in hex string format, along with its transaction id.
func FinalizeAndExtractTransaction(args FinalizeAndExtractTransactionArgs) (string, string, error) {
	if err := args.validate(); err != nil {
		return "", "", err
	}

	ptx, _ := psetv2.NewPsetFromBase64(args.PsetBase64)

	ok, err := ptx.ValidateAllSignatures()
	if err != nil {
		return "", "", err
	}
	if !ok {
		return "", "", ErrInvalidSignatures
	}

	if err := psetv2.FinalizeAll(ptx); err != nil {
		return "", "", err
	}

	tx, err := psetv2.Extract(ptx)
	if err != nil {
		return "", "", err
	}
	txHex, err := tx.ToHex()
	if err != nil {
		return "", "", err
	}
	return txHex, tx.TxHash().String(), nil
}
