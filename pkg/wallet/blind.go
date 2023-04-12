package wallet

import (
	"encoding/hex"
	"fmt"

	"github.com/vulpemventures/go-elements/confidential"
	"github.com/vulpemventures/go-elements/elementsutil"
	"github.com/vulpemventures/go-elements/psetv2"
)

var (
	ErrMissingOwnedInputs       = fmt.Errorf("missing list of owned inputs")
	ErrMissingBlindingMasterKey = fmt.Errorf("missing blinding master key")
	ErrBlindInvalidInputIndex   = fmt.Errorf("input index to blind is out of range")
)

type BlindPsetWithOwnedInputsArgs struct {
	PsetBase64         string
	OwnedInputsByIndex map[uint32]Input
	LastBlinder        bool
}

func (a BlindPsetWithOwnedInputsArgs) validate() error {
	if a.PsetBase64 == "" {
		return ErrMissingPset
	}
	ptx, err := psetv2.NewPsetFromBase64(a.PsetBase64)
	if err != nil {
		return err
	}
	if len(a.OwnedInputsByIndex) == 0 {
		return ErrMissingOwnedInputs
	}
	for i, in := range a.OwnedInputsByIndex {
		if int(i) >= int(ptx.Global.InputCount) {
			return ErrBlindInvalidInputIndex
		}
		if err := in.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (a BlindPsetWithOwnedInputsArgs) ownedInputs() map[uint32]psetv2.OwnedInput {
	ownedInputs := make(map[uint32]psetv2.OwnedInput)
	for i, in := range a.OwnedInputsByIndex {
		ownedInputs[i] = psetv2.OwnedInput{
			Index:        i,
			Value:        in.Value,
			Asset:        in.Asset,
			ValueBlinder: in.ValueBlinder,
			AssetBlinder: in.AssetBlinder,
		}
	}
	return ownedInputs
}

func (a BlindPsetWithOwnedInputsArgs) inputIndexes() []uint32 {
	ownedInputIndexes := make([]uint32, 0, len(a.OwnedInputsByIndex))
	for i := range a.OwnedInputsByIndex {
		ownedInputIndexes = append(ownedInputIndexes, i)
	}
	return ownedInputIndexes
}

func BlindPsetWithOwnedInputs(
	args BlindPsetWithOwnedInputsArgs,
) (string, error) {
	if err := args.validate(); err != nil {
		return "", err
	}

	ptx, _ := psetv2.NewPsetFromBase64(args.PsetBase64)
	if !ptx.NeedsBlinding() {
		return args.PsetBase64, nil
	}

	ownedInputIndexes := args.inputIndexes()
	ownedInputsByIndex := args.ownedInputs()
	inputIndexes := ownedInputIndexes

	blindingValidator := confidential.NewZKPValidator()
	blindingGenerator, err := confidential.NewZKPGeneratorFromOwnedInputs(
		ownedInputsByIndex, nil,
	)
	if err != nil {
		return "", err
	}

	outputIndexesToBlind := getOutputIndexesToBlind(ptx, inputIndexes)
	ownedInputs, err := blindingGenerator.UnblindInputs(ptx, inputIndexes)
	if err != nil {
		return "", err
	}
	outBlindArgs, err := blindingGenerator.BlindOutputs(
		ptx, outputIndexesToBlind,
	)
	if err != nil {
		return "", err
	}

	blinder, err := psetv2.NewBlinder(
		ptx, ownedInputs, blindingValidator, blindingGenerator,
	)
	if err != nil {
		return "", err
	}

	blindingFn := blinder.BlindNonLast
	if args.LastBlinder {
		blindingFn = blinder.BlindLast
	}
	if err := blindingFn(nil, outBlindArgs); err != nil {
		return "", err
	}

	return ptx.ToBase64()
}

type BlindPsetWithMasterKeyArgs struct {
	PsetBase64        string
	BlindingMasterKey []byte
	ExtraBlindingKeys map[string][]byte
	LastBlinder       bool
}

func (a BlindPsetWithMasterKeyArgs) validate() error {
	if a.PsetBase64 == "" {
		return ErrMissingPset
	}
	if _, err := psetv2.NewPsetFromBase64(a.PsetBase64); err != nil {
		return err
	}
	if len(a.BlindingMasterKey) <= 0 {
		return ErrMissingBlindingMasterKey
	}
	return nil
}

func (a BlindPsetWithMasterKeyArgs) inputIndexes() ([]uint32, []uint32) {
	ptx, _ := psetv2.NewPsetFromBase64(a.PsetBase64)
	if len(a.ExtraBlindingKeys) == 0 {
		indexes := make([]uint32, 0, len(ptx.Inputs))
		for i := range ptx.Inputs {
			indexes = append(indexes, uint32(i))
		}
		return indexes, nil
	}

	ownedIns, notOwnedIns := make([]uint32, 0), make([]uint32, 0)
	for i, in := range ptx.Inputs {
		prevout := in.GetUtxo()
		prevoutScript := hex.EncodeToString(prevout.Script)
		if _, ok := a.ExtraBlindingKeys[prevoutScript]; ok {
			notOwnedIns = append(notOwnedIns, uint32(i))
			continue
		}
		ownedIns = append(ownedIns, uint32(i))
	}
	return ownedIns, notOwnedIns
}

func BlindPsetWithMasterKey(
	args BlindPsetWithMasterKeyArgs,
) (string, error) {
	if err := args.validate(); err != nil {
		return "", err
	}

	ptx, _ := psetv2.NewPsetFromBase64(args.PsetBase64)
	if !ptx.NeedsBlinding() {
		return args.PsetBase64, nil
	}

	ownedInputIndexes, notOwnedInputIndexes := args.inputIndexes()
	extraInputs, err := unblindNotOwnedInputs(
		ptx, args.ExtraBlindingKeys, notOwnedInputIndexes,
	)
	if err != nil {
		return "", err
	}

	blindingValidator := confidential.NewZKPValidator()
	blindingGenerator, err := confidential.NewZKPGeneratorFromMasterBlindingKey(
		args.BlindingMasterKey, nil,
	)
	if err != nil {
		return "", err
	}

	ownedInputs, err := blindingGenerator.UnblindInputs(ptx, ownedInputIndexes)
	if err != nil {
		return "", err
	}

	inputIndexes := ownedInputIndexes
	if len(args.ExtraBlindingKeys) > 0 {
		inputIndexes = append(inputIndexes, notOwnedInputIndexes...)
	}

	outputIndexesToBlind := getOutputIndexesToBlind(
		ptx, inputIndexes,
	)

	outBlindArgs, err := blindingGenerator.BlindOutputs(
		ptx, outputIndexesToBlind,
	)
	if err != nil {
		return "", err
	}

	// extraInputs is empty in case no ExtraBlindingKeys are passaed within args.
	ownedInputs = append(ownedInputs, extraInputs...)
	blinder, err := psetv2.NewBlinder(
		ptx, ownedInputs, blindingValidator, blindingGenerator,
	)
	if err != nil {
		return "", err
	}

	blindingFn := blinder.BlindNonLast
	if args.LastBlinder {
		blindingFn = blinder.BlindLast
	}
	if err := blindingFn(nil, outBlindArgs); err != nil {
		return "", err
	}

	return ptx.ToBase64()
}

func unblindNotOwnedInputs(
	ptx *psetv2.Pset, blindKeysByScript map[string][]byte, inputIndexes []uint32,
) ([]psetv2.OwnedInput, error) {
	if len(blindKeysByScript) == 0 {
		return nil, nil
	}

	revealedOuts := make([]psetv2.OwnedInput, 0)
	for _, inIndex := range inputIndexes {
		in := ptx.Inputs[inIndex]
		prevout := in.GetUtxo()
		script := hex.EncodeToString(prevout.Script)
		key := blindKeysByScript[script]
		revealed, err := confidential.UnblindOutputWithKey(prevout, key)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to unblind not owned input %d with given blinding key "+
					"generated by prevout script %s", inIndex, script,
			)
		}
		revealedOuts = append(revealedOuts, psetv2.OwnedInput{
			Value:        revealed.Value,
			Asset:        hex.EncodeToString(elementsutil.ReverseBytes(revealed.Asset)),
			ValueBlinder: revealed.ValueBlindingFactor,
			AssetBlinder: revealed.AssetBlindingFactor,
		})
	}
	return revealedOuts, nil
}

func getOutputIndexesToBlind(
	ptx *psetv2.Pset, ownedInputIndexes []uint32,
) []uint32 {
	ownedOuts := make([]uint32, 0)

	isOwnedOutput := func(index uint32) bool {
		for _, i := range ownedInputIndexes {
			if i == index {
				return true
			}
		}
		return false
	}

	for i, out := range ptx.Outputs {
		if out.NeedsBlinding() {
			if !isOwnedOutput(out.BlinderIndex) {
				continue
			}
			ownedOuts = append(ownedOuts, uint32(i))
		}
	}
	return ownedOuts
}
