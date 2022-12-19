package wallet

import (
	"github.com/btcsuite/btcd/txscript"
)

var (
	scripsigtSizeByScriptType = map[int]int{
		P2PK:        140, // len + opcode + sig + opcode + pubkey uncompressed
		P2PKH:       108, // len + opcode + sig + opcode + pubkey
		P2SH_P2WPKH: 23,  // len + p2wpkh script
		P2SH_P2WSH:  35,  // len + p2wsh script
		P2WPKH:      1,   // no scriptsig, still len is serialized
		P2WSH:       1,   // no scriptsig
	}
)

// EstimateTxSize makes an estimation of the virtual size of a transaction for
// which is required to specify the type of the inputs and outputs according to
// those of the Bitcoin standard (P2PK, P2PKH, P2MS, P2SH(P2WPKH), P2SH(P2WSH),
// P2WPKH, P2WSH).
// The estimation might not be accurate in case of one or more P2MS inputs
// since the method is not able to retrieve the size of redeem script containg
// all pubkeys, nor it expects anyone as arg.
func EstimateTxSize(inputs []Input, outputs []Output) uint64 {
	inScriptsigsSize, inWitnessesSize := make([]int, 0), make([]int, 0)
	for _, in := range inputs {
		inType := in.ScriptType()
		scriptsigSize := in.ScriptSigSize
		witnessSize := in.WitnessSize
		if scriptsigSize <= 0 {
			scriptsigSize = scripsigtSizeByScriptType[inType]
		}
		if witnessSize <= 0 {
			if len(in.RedeemScript) > 0 {
				_, m, _ := txscript.CalcMultiSigStats(in.RedeemScript)
				// num of sigs + separators + size of redeem script
				witnessSize = 75*m + m - 1 + varSliceSerializeSize(in.RedeemScript)
			} else {
				// len + witness[sig,pubkey]
				witnessSize = (1 + 107)
			}
		}
		// add no issuance proof + no token proof + no pegin
		witnessSize += 1 + 1 + 1
		inScriptsigsSize = append(inScriptsigsSize, scriptsigSize)
		inWitnessesSize = append(inWitnessesSize, witnessSize)
	}
	outsSize, outWitnessesSize := make([]int, 0), make([]int, 0)
	for _, out := range outputs {
		// no rangeproof + no surjectionproof
		witnessSize := 1 + 1
		// asset + amount + empty noce
		outSize := 33 + 9 + 1
		if out.IsConfidential() {
			outSize = 33 + 33 + 33
			// size(rangeproof) + proof + size(sujectionproof) + proof
			witnessSize = (3 + 4174 + 1 + 131)
		}
		outsSize = append(outsSize, outSize+out.ScriptSize())
		outWitnessesSize = append(outWitnessesSize, witnessSize)
	}

	txSize := estimateTxSize(
		inScriptsigsSize, inWitnessesSize, outsSize, outWitnessesSize,
	)
	return uint64(txSize)
}

// EstimateFees estimates the virtual size of the transaciton composed of the
// given Inputs and Outputs and then returns the corresponding fee amount based
// on the given mSats/Byte ratio.
func EstimateFees(
	inputs []Input, outputs []Output, millisatsPerByte uint64,
) uint64 {
	txSize := EstimateTxSize(inputs, outputs)
	satsPerByte := float64(millisatsPerByte) / 1000
	return uint64(float64(txSize) * satsPerByte)
}

func estimateTxSize(
	inScripsigsSize, inWitnessesSize, outsSize, outWitnessesSize []int,
) int {
	baseSize := calcTxSize(
		false, inScripsigsSize, inWitnessesSize, outsSize, outWitnessesSize,
	)
	totalSize := calcTxSize(
		true, inScripsigsSize, inWitnessesSize, outsSize, outWitnessesSize,
	)

	weight := baseSize*3 + totalSize
	vsize := (weight + 3) / 4

	return vsize
}

func calcTxSize(
	withWitness bool,
	inScripsigsSize, inWitnessesSize, outsSize, outWitnessesSize []int,
) int {
	txSize := calcTxBaseSize(inScripsigsSize, outsSize)
	if withWitness {
		txSize += calcTxWitnessSize(inWitnessesSize, outWitnessesSize)
	}
	return txSize
}

func calcTxBaseSize(
	inScripsigsSize, outNonWitnessesSize []int,
) int {
	// hash + index + sequence
	inBaseSize := 40
	insSize := 0
	for _, scriptSigSize := range inScripsigsSize {
		insSize += inBaseSize + scriptSigSize
	}

	outsSize := 0
	for _, outSize := range outNonWitnessesSize {
		outsSize += outSize
	}
	// size of unconf fee out
	// asset + unconf value + empty script + empty nonce
	outsSize += 33 + 9 + 1 + 1

	return 9 +
		varIntSerializeSize(uint64(len(inScripsigsSize))) +
		varIntSerializeSize(uint64(len(outNonWitnessesSize)+1)) +
		insSize + outsSize
}

func calcTxWitnessSize(
	inWitnessesSize, outWitnessesSize []int,
) int {
	insSize := 0
	for _, witnessSize := range inWitnessesSize {
		insSize += witnessSize
	}

	outsSize := 0
	for _, witnessSize := range outWitnessesSize {
		outsSize += witnessSize
	}
	// size of proofs for unconf fee out
	outsSize += 1 + 1

	return insSize + outsSize
}
