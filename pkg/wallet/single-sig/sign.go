package singlesig

import (
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/vulpemventures/go-elements/elementsutil"
	"github.com/vulpemventures/go-elements/payment"
	"github.com/vulpemventures/go-elements/psetv2"
	"github.com/vulpemventures/go-elements/taproot"
	"github.com/vulpemventures/go-elements/transaction"
	"github.com/vulpemventures/ocean/pkg/wallet"
	path "github.com/vulpemventures/ocean/pkg/wallet/derivation-path"
)

type SignTransactionArgs struct {
	TxHex        string
	InputsToSign map[uint32]wallet.Input
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
		derivationPath, err := path.ParseDerivationPath(in.DerivationPath)
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
	if _, err := psetv2.NewPsetFromBase64(a.PsetBase64); err != nil {
		return err
	}
	if len(a.DerivationPathMap) <= 0 {
		return ErrMissingDerivationPaths
	}

	for script, pathStr := range a.DerivationPathMap {
		derivationPath, err := path.ParseDerivationPath(pathStr)
		if err != nil {
			return fmt.Errorf(
				"invalid derivation path '%s' for script '%s': %v",
				pathStr, script, err,
			)
		}
		err = checkDerivationPath(derivationPath)
		if err != nil {
			return fmt.Errorf(
				"invalid derivation path '%s' for script '%s': %v",
				pathStr, script, err,
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
		path, ok := args.DerivationPathMap[hex.EncodeToString(in.GetUtxo().Script)]
		if ok {
			err := w.signInput(ptx, i, path, args.sighashType())
			if err != nil {
				return "", err
			}
		}
	}

	return ptx.ToBase64()
}

type SignTaprootArgs struct {
	PsetBase64        string
	DerivationPathMap map[string]string
	GenesisBlockHash  string
	SighashType       txscript.SigHashType
}

func (a SignTaprootArgs) validate() error {
	if _, err := psetv2.NewPsetFromBase64(a.PsetBase64); err != nil {
		return err
	}
	if len(a.DerivationPathMap) <= 0 {
		return ErrMissingDerivationPaths
	}
	if len(a.GenesisBlockHash) <= 0 {
		return fmt.Errorf("missing genesis block hash")
	}
	if _, err := chainhash.NewHashFromStr(a.GenesisBlockHash); err != nil {
		return fmt.Errorf("invalid genesis block hash: %s", err)
	}

	for script, pathStr := range a.DerivationPathMap {
		derivationPath, err := path.ParseDerivationPath(pathStr)
		if err != nil {
			return fmt.Errorf(
				"invalid derivation path '%s' for script '%s': %v",
				pathStr, script, err,
			)
		}
		err = checkDerivationPath(derivationPath)
		if err != nil {
			return fmt.Errorf(
				"invalid derivation path '%s' for script '%s': %v",
				pathStr, script, err,
			)
		}
	}

	return nil
}

func (a SignTaprootArgs) genesisBlockHash() *chainhash.Hash {
	hash, _ := chainhash.NewHashFromStr(a.GenesisBlockHash)
	return hash
}

func (w *Wallet) SignTaproot(args SignTaprootArgs) (string, error) {
	if err := args.validate(); err != nil {
		return "", err
	}
	if err := w.validate(); err != nil {
		return "", err
	}

	ptx, _ := psetv2.NewPsetFromBase64(args.PsetBase64)
	for i, in := range ptx.Inputs {
		path, ok := args.DerivationPathMap[hex.EncodeToString(in.GetUtxo().Script)]
		if ok {
			err := w.signTaprootInput(ptx, i, path, args.SighashType, args.genesisBlockHash())
			if err != nil {
				return "", err
			}
		}
	}

	return ptx.ToBase64()
}

func (w *Wallet) signTxInput(
	tx *transaction.Transaction, inIndex uint32, input wallet.Input,
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
		if err := signer.AddInSighashType(inIndex, sighashType); err != nil {
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
		input.GetUtxo().Script, nil, nil,
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
		inIndex, script, ptx.Inputs[inIndex].GetUtxo().Value, input.SigHashType,
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

func (w *Wallet) signTaprootInput(
	ptx *psetv2.Pset, inIndex int, derivationPath string,
	sighashType txscript.SigHashType, genesisBlockHash *chainhash.Hash,
) error {
	signer, err := psetv2.NewSigner(ptx)
	if err != nil {
		return err
	}
	input := ptx.Inputs[inIndex]
	if err := signer.AddInSighashType(inIndex, sighashType); err != nil {
		return err
	}

	prvkey, pubkey, err := w.DeriveSigningKeyPair(DeriveSigningKeyPairArgs{
		DerivationPath: derivationPath,
	})
	if err != nil {
		return err
	}

	for _, derivation := range input.TapBip32Derivation {
		for _, hash := range derivation.LeafHashes {
			leafHash, err := chainhash.NewHash(hash)
			if err != nil {
				return err
			}

			tapScriptSig, err := signTaproot(
				ptx, inIndex, prvkey, pubkey, sighashType, genesisBlockHash, leafHash,
			)
			if err != nil {
				return err
			}

			if err := signer.SignTaprootInputTapscriptSig(inIndex, *tapScriptSig); err != nil {
				return err
			}
		}
		// If there are no leaf hashes, sign as keypath
		if len(derivation.LeafHashes) <= 0 {
			tweakedPrvKey := taproot.TweakTaprootPrivKey(prvkey, input.TapMerkleRoot)
			tweakedPubKey := taproot.ComputeTaprootOutputKey(pubkey, input.TapMerkleRoot)

			tapScriptSig, err := signTaproot(
				ptx, inIndex, tweakedPrvKey, tweakedPubKey, sighashType, genesisBlockHash, nil,
			)
			if err != nil {
				return err
			}
			if err := signer.SignTaprootInputTapscriptSig(inIndex, *tapScriptSig); err != nil {
				return err
			}
		}
	}

	return nil
}

func signTaproot(
	ptx *psetv2.Pset, inIndex int,
	prvkey *btcec.PrivateKey, pubkey *btcec.PublicKey,
	sighashType txscript.SigHashType, genesisBlockHash, leafHash *chainhash.Hash,
) (*psetv2.TapScriptSig, error) {
	unsignedTx, err := ptx.UnsignedTx()
	if err != nil {
		return nil, err
	}

	input := ptx.Inputs[inIndex]
	prevoutScripts := [][]byte{input.GetUtxo().Script}
	prevoutAssets := [][]byte{input.GetUtxo().Asset}
	prevoutValues := [][]byte{input.GetUtxo().Value}

	hashForSignature := unsignedTx.HashForWitnessV1(
		inIndex, prevoutScripts, prevoutAssets, prevoutValues, sighashType, genesisBlockHash, leafHash, nil,
	)
	signature, err := schnorr.Sign(prvkey, hashForSignature[:])
	if err != nil {
		return nil, fmt.Errorf("failed to sign with schnorr: %s", err)
	}
	if !signature.Verify(hashForSignature[:], pubkey) {
		return nil, fmt.Errorf("signature verification failed for input %d", inIndex)
	}

	sig := signature.Serialize()
	if sighashType != txscript.SigHashDefault {
		sig = append(sig, byte(sighashType))
	}

	return &psetv2.TapScriptSig{
		PartialSig: psetv2.PartialSig{
			PubKey:    schnorr.SerializePubKey(pubkey),
			Signature: sig,
		},
		LeafHash: leafHash[:],
	}, nil
}
