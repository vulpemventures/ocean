package wallet

import (
	"fmt"

	"github.com/vulpemventures/go-elements/address"
	"github.com/vulpemventures/go-elements/elementsutil"
	"github.com/vulpemventures/go-elements/transaction"
)

const (
	P2PK = iota
	P2PKH
	P2MS
	P2SH_P2WPKH
	P2SH_P2WSH
	P2WPKH
	P2WSH
)

var (
	scriptTypes = map[int]int{
		address.P2PkhScript:  P2PKH,
		address.P2ShScript:   P2SH_P2WPKH,
		address.P2WpkhScript: P2WPKH,
		address.P2WshScript:  P2WSH,
	}
)

var (
	ErrInputMissingTxid = fmt.Errorf("input is missing txid")
	ErrInputInvalidTxid = fmt.Errorf("invalid input txid length: must be exactly 32 bytes")
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
	RedeemScript    []byte
	ScriptSigSize   int
	WitnessSize     int
}

func (i Input) Validate() error {
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

func (i Input) Prevout() *transaction.TxOutput {
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
		Asset:  asset,
		Value:  value,
		Script: i.Script,
		Nonce:  nonce,
	}
}

func (i Input) ScriptType() int {
	t := scriptTypes[address.GetScriptType(i.Script)]
	if t == P2SH_P2WPKH && len(i.RedeemScript) > 0 {
		t = P2SH_P2WSH
	}
	return t
}
