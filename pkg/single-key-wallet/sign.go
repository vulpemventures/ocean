package wallet

import (
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"github.com/btcsuite/btcd/txscript"
	"github.com/vulpemventures/go-elements/elementsutil"
	"github.com/vulpemventures/go-elements/payment"
	"github.com/vulpemventures/go-elements/psetv2"
	"github.com/vulpemventures/go-elements/transaction"
)

type SignTransactionArgs struct {
	TxHex        string
	InputsToSign map[uint32]Input
	SigHashType  txscript.SigHashType
}

func (a SignTransactionArgs) validate() error {
	tx, err := transaction.NewTxFromHex(a.TxHex)
	if err != nil {
		return err
	}
	if len(a.InputsToSign) <= 0 {
		return ErrMissingPrevOuts
	}

	for index, in := range a.InputsToSign {
		if int(index) >= len(tx.Inputs) {
			return fmt.Errorf("input index %d out of range", index)
		}
		if in.DerivationPath == "" {
			return fmt.Errorf(
				"invalid input %d: %s", index, ErrMissingDerivationPath,
			)
		}
		derivationPath, err := ParseDerivationPath(in.DerivationPath)
		if err != nil {
			return fmt.Errorf(
				"invalid derivation path '%s' for input %d: %v",
				in.DerivationPath, index, err,
			)
		}
		if err = checkDerivationPath(derivationPath); err != nil {
			return fmt.Errorf(
				"invalid derivation path '%s' for input %d: %v",
				in.DerivationPath, index, err,
			)
		}
	}

	return nil
}

func (a SignTransactionArgs) sighashType() txscript.SigHashType {
	if a.SigHashType == 0 {
		return txscript.SigHashAll
	}
	return a.SigHashType
}

// SignTransaction signs all requested inputs of the given raw transaction.
func (w *Wallet) SignTransaction(args SignTransactionArgs) (string, error) {
	if err := args.validate(); err != nil {
		return "", err
	}
	if err := w.validate(); err != nil {
		return "", err
	}

	tx, _ := transaction.NewTxFromHex(args.TxHex)
	for i, in := range args.InputsToSign {
		if err := w.signTxInput(tx, i, in, args.sighashType()); err != nil {
			return "", err
		}
	}

	return tx.ToHex()
}

type SignPsetArgs struct {
	PsetBase64        string
	DerivationPathMap map[string]string
	SigHashType       txscript.SigHashType
}

func (a SignPsetArgs) validate() error {
	ptx, err := psetv2.NewPsetFromBase64(a.PsetBase64)
	if err != nil {
		return err
	}
	if len(a.DerivationPathMap) <= 0 {
		return ErrMissingDerivationPaths
	}

	for script, path := range a.DerivationPathMap {
		derivationPath, err := ParseDerivationPath(path)
		if err != nil {
			return fmt.Errorf(
				"invalid derivation path '%s' for script '%s': %v",
				path, script, err,
			)
		}
		err = checkDerivationPath(derivationPath)
		if err != nil {
			return fmt.Errorf(
				"invalid derivation path '%s' for script '%s': %v",
				path, script, err,
			)
		}
	}

	for i, in := range ptx.Inputs {
		script := in.GetUtxo().Script
		_, ok := a.DerivationPathMap[hex.EncodeToString(script)]
		if !ok {
			return fmt.Errorf(
				"derivation path not found in list for input %d with script '%s'",
				i,
				hex.EncodeToString(script),
			)
		}
	}

	return nil
}

func (a SignPsetArgs) sighashType() txscript.SigHashType {
	if a.SigHashType == 0 {
		return txscript.SigHashAll
	}
	return a.SigHashType
}

// SignPset signs all inputs of a partial transaction matching the given
// scripts of the derivation path map.
func (w *Wallet) SignPset(args SignPsetArgs) (string, error) {
	if err := args.validate(); err != nil {
		return "", err
	}
	if err := w.validate(); err != nil {
		return "", err
	}

	ptx, _ := psetv2.NewPsetFromBase64(args.PsetBase64)
	for i, in := range ptx.Inputs {
		path := args.DerivationPathMap[hex.EncodeToString(in.WitnessUtxo.Script)]
		err := w.signInput(ptx, i, path, args.sighashType())
		if err != nil {
			return "", err
		}
	}

	return ptx.ToBase64()
}

func (w *Wallet) signTxInput(
	tx *transaction.Transaction, inIndex uint32, input Input,
	sighashType txscript.SigHashType,
) error {
	prvkey, pubkey, err := w.DeriveSigningKeyPair(DeriveSigningKeyPairArgs{
		DerivationPath: input.DerivationPath,
	})
	if err != nil {
		return err
	}

	pay, err := payment.FromScript(input.Script, nil, nil)
	if err != nil {
		return err
	}
	script := pay.Script

	unsignedTx := tx.Copy()
	for _, in := range unsignedTx.Inputs {
		in.Script = nil
		in.Witness = nil
		in.PeginWitness = nil
		in.InflationRangeProof = nil
		in.IssuanceRangeProof = nil
	}

	var value []byte
	if len(input.ValueCommitment) == 0 {
		value, _ = elementsutil.ValueToBytes(input.Value)
	}
	hashForSignature := unsignedTx.HashForWitnessV0(
		int(inIndex), script, value, sighashType,
	)

	signature := ecdsa.Sign(prvkey, hashForSignature[:])
	if !signature.Verify(hashForSignature[:], pubkey) {
		return fmt.Errorf(
			"signature verification failed for input %d",
			inIndex,
		)
	}

	sigWithSigHashType := append(signature.Serialize(), byte(sighashType))

	tx.Inputs[inIndex].Witness = transaction.TxWitness{
		sigWithSigHashType, pubkey.SerializeCompressed(),
	}

	return nil
}

func (w *Wallet) signInput(
	ptx *psetv2.Pset, inIndex int, derivationPath string,
	sighashType txscript.SigHashType,
) error {
	signer, err := psetv2.NewSigner(ptx)
	if err != nil {
		return err
	}

	if ptx.Inputs[inIndex].SigHashType == 0 {
		if err := signer.AddInSighashType(sighashType, inIndex); err != nil {
			return err
		}
	}
	input := ptx.Inputs[inIndex]

	prvkey, pubkey, err := w.DeriveSigningKeyPair(DeriveSigningKeyPairArgs{
		DerivationPath: derivationPath,
	})

	if err != nil {
		return err
	}

	pay, err := payment.FromScript(
		input.WitnessUtxo.Script, nil, nil,
	)
	if err != nil {
		return err
	}

	script := pay.Script
	unsingedTx, err := ptx.UnsignedTx()
	if err != nil {
		return err
	}

	hashForSignature := unsingedTx.HashForWitnessV0(
		inIndex, script, ptx.Inputs[inIndex].WitnessUtxo.Value, input.SigHashType,
	)

	signature := ecdsa.Sign(prvkey, hashForSignature[:])
	if !signature.Verify(hashForSignature[:], pubkey) {
		return fmt.Errorf(
			"signature verification failed for input %d",
			inIndex,
		)
	}

	sigWithSigHashType := append(signature.Serialize(), byte(input.SigHashType))
	return signer.SignInput(
		inIndex, sigWithSigHashType, pubkey.SerializeCompressed(), nil, nil,
	)
}
