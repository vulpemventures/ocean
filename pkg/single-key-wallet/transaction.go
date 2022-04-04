package wallet

import (
	"fmt"

	"github.com/vulpemventures/go-elements/address"
	"github.com/vulpemventures/go-elements/elementsutil"
	"github.com/vulpemventures/go-elements/psetv2"
	"github.com/vulpemventures/go-elements/transaction"
)

var (
	// DummyFeeAmount is used as the fee amount to cover when coin-selecting the
	// inputs to use to cover the true fee amount, which, instead, is calculated
	// with more precision from the tx size.
	// The real fee amount strictly depends on the number of tx inputs and
	// outputs, and even input types.
	// This value is thought for transactions on TDEX network, whose are composed
	// by at least 3 inputs and 6 outputs.
	// If all inputs are wrapped or native segwit, is shouls be unlikely for the
	// tx virtual size to be higher than 700 vB/sat, taking into account that
	// this pkg supports ONLY native segwit scripts/addresses.
	// For any other case this value can be tweaked at will.
	DummyFeeAmount uint64 = 700
)

// Input is the data structure representing an input to be added to a partial
// transaction, therefore including the previous outpoint as long as all the
// info about the prevout itself (and the derivation path generating its script
// as extra info).
type Input struct {
	TxID            string
	TxIndex         uint32
	Value           uint64
	Asset           string
	Script          []byte
	ValueBlinder    []byte
	AssetBlinder    []byte
	ValueCommitment []byte
	AssetCommitment []byte
	Nonce           []byte
	RangeProof      []byte
	SurjectionProof []byte
	DerivationPath  string
}

func (i Input) validate() error {
	if i.TxID == "" {
		return ErrInputMissingTxid
	}
	buf, err := elementsutil.TxIDToBytes(i.TxID)
	if err != nil {
		return err
	}
	if len(buf) != 32 {
		return ErrInputInvalidTxid
	}
	return nil
}

func (i Input) prevout() *transaction.TxOutput {
	value := i.ValueCommitment
	if len(value) == 0 {
		value, _ = elementsutil.ValueToBytes(i.Value)
	}
	asset := i.AssetCommitment
	if len(asset) == 0 {
		asset, _ = elementsutil.AssetHashToBytes(i.Asset)
	}
	nonce := i.Nonce
	if len(nonce) == 0 {
		nonce = make([]byte, 1)
	}
	return &transaction.TxOutput{
		Asset:           asset,
		Value:           value,
		Script:          i.Script,
		Nonce:           nonce,
		RangeProof:      i.RangeProof,
		SurjectionProof: i.SurjectionProof,
	}
}

func (i Input) scriptType() int {
	return scriptTypes[address.GetScriptType(i.Script)]
}

// Output is the data structure representing an output to be added to a partial
// transaction, therefore inclusing asset, amount and address.
type Output struct {
	Asset   string
	Amount  uint64
	Address string
}

func (o Output) validate() error {
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
	if o.Address != "" {
		if _, err := address.IsConfidential(o.Address); err != nil {
			return err
		}
	}
	return nil
}

func (o Output) scriptType() int {
	if o.Address == "" {
		return -1
	}
	script, _ := address.ToOutputScript(o.Address)
	return scriptTypes[address.GetScriptType(script)]
}

type CreatePsetArgs struct {
	Inputs  []Input
	Outputs []Output
}

func (a CreatePsetArgs) validate() error {
	for i, in := range a.Inputs {
		if err := in.validate(); err != nil {
			return fmt.Errorf("invalid input %d: %s", i, err)
		}
	}

	for i, out := range a.Outputs {
		if err := out.validate(); err != nil {
			return fmt.Errorf("invalid output %d: %s", i, err)
		}
		isConfidential, _ := address.IsConfidential(out.Address)
		if isConfidential && len(a.Inputs) == 0 {
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
		outs = append(outs, psetv2.OutputArgs{
			Amount:  out.Amount,
			Asset:   out.Asset,
			Address: out.Address,
		})
	}
	return outs
}

// CreatePset creates a new partial transaction with given inputs and outputs.
func (w *Wallet) CreatePset(args CreatePsetArgs) (string, error) {
	if err := args.validate(); err != nil {
		return "", err
	}

	ptx, err := psetv2.New(args.inputs(), args.outputs(), 0)
	if err != nil {
		return "", err
	}

	updater, err := psetv2.NewUpdater(ptx)
	if err != nil {
		return "", err
	}
	for i, in := range args.Inputs {
		updater.AddInWitnessUtxo(in.prevout(), i)
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
		if err := in.validate(); err != nil {
			return fmt.Errorf("invalid input %d: %s", i, err)
		}
	}

	for i, out := range a.Outputs {
		if err := out.validate(); err != nil {
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
		outs = append(outs, psetv2.OutputArgs{
			Amount:       out.Amount,
			Asset:        out.Asset,
			Address:      out.Address,
			BlinderIndex: blinderIndex,
		})
	}
	return outs
}

// UpdatesPset adds inputs and outputs to the given partial transaction.
func (w *Wallet) UpdatePset(args UpdatePsetArgs) (string, error) {
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
		updater.AddInWitnessUtxo(in.prevout(), int(nextInputIndex)+i)
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
