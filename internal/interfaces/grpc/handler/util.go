package grpc_handler

import (
	"encoding/hex"
	"fmt"

	"github.com/vulpemventures/go-elements/address"
	"github.com/vulpemventures/go-elements/elementsutil"
	pb "github.com/vulpemventures/ocean/api-spec/protobuf/gen/go/ocean/v1"
	"github.com/vulpemventures/ocean/internal/core/application"
	"github.com/vulpemventures/ocean/internal/core/domain"
)

func parseMnemonic(mnemonic string) (string, error) {
	if mnemonic == "" {
		return "", fmt.Errorf("missing mnemonic")
	}
	return mnemonic, nil
}

func parsePassword(password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("missing password")
	}
	return password, nil
}

func parseNetwork(network string) pb.GetInfoResponse_Network {
	switch network {
	case "liquid":
		return pb.GetInfoResponse_NETWORK_MAINNET
	case "testnet":
		return pb.GetInfoResponse_NETWORK_TESTNET
	case "regtest":
		return pb.GetInfoResponse_NETWORK_REGTEST
	default:
		return pb.GetInfoResponse_NETWORK_UNSPECIFIED
	}
}

func parseAccounts(accounts []application.AccountInfo) []*pb.AccountInfo {
	list := make([]*pb.AccountInfo, 0, len(accounts))
	for _, a := range accounts {
		list = append(list, &pb.AccountInfo{
			Name:              a.Key.Name,
			Index:             a.Key.Index,
			Xpubs:             []string{a.Xpub},
			DerivationPath:    a.DerivationPath,
			MasterBlindingKey: a.MasterBlindingKey,
		})
	}
	return list
}

func parseAccountName(name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("missing account name")
	}
	return name, nil
}

func parseUtxos(utxos []domain.UtxoInfo) []*pb.Utxo {
	list := make([]*pb.Utxo, 0, len(utxos))
	for _, u := range utxos {
		emptyStatus := domain.UtxoStatus{}
		var spentStatus, confirmedStatus *pb.UtxoStatus
		if u.SpentStatus != emptyStatus {
			spentStatus = &pb.UtxoStatus{
				Txid: u.SpentStatus.Txid,
				BlockInfo: &pb.BlockDetails{
					Hash:      u.SpentStatus.BlockHash,
					Height:    u.SpentStatus.BlockHeight,
					Timestamp: u.SpentStatus.BlockTime,
				},
			}
		}
		if u.ConfirmedStatus != emptyStatus {
			confirmedStatus = &pb.UtxoStatus{
				BlockInfo: &pb.BlockDetails{
					Hash:      u.ConfirmedStatus.BlockHash,
					Height:    u.ConfirmedStatus.BlockHeight,
					Timestamp: u.ConfirmedStatus.BlockTime,
				},
			}
		}
		list = append(list, &pb.Utxo{
			Txid:            u.Key().TxID,
			Index:           u.Key().VOut,
			Asset:           u.Asset,
			Value:           u.Value,
			Script:          hex.EncodeToString(u.Script),
			AssetBlinder:    elementsutil.TxIDFromBytes(u.AssetBlinder),
			ValueBlinder:    elementsutil.TxIDFromBytes(u.ValueBlinder),
			AccountName:     u.AccountName,
			SpentStatus:     spentStatus,
			ConfirmedStatus: confirmedStatus,
		})
	}
	return list
}

func parseBlockDetails(tx application.TransactionInfo) *pb.BlockDetails {
	if tx.BlockHash == "" {
		return nil
	}
	return &pb.BlockDetails{
		Hash:      tx.BlockHash,
		Height:    tx.BlockHeight,
		Timestamp: int64(tx.BlockHeight),
	}
}

func parseInputs(ins []*pb.Input) ([]application.Input, error) {
	inputs := make([]application.Input, 0, len(ins))
	for _, in := range ins {
		inputs = append(inputs, application.Input{
			TxID:          in.GetTxid(),
			VOut:          in.GetIndex(),
			Script:        in.GetScript(),
			ScriptSigSize: int(in.GetScriptsigSize()),
			WitnessSize:   int(in.GetWitnessSize()),
		})
	}
	return inputs, nil
}

func parseOutputs(outs []*pb.Output) ([]application.Output, error) {
	outputs := make([]application.Output, 0, len(outs))
	for _, out := range outs {
		var script, blindKey []byte
		if addr := out.GetAddress(); addr != "" {
			isConf, err := address.IsConfidential(addr)
			if err != nil {
				return nil, err
			}
			if isConf {
				res, _ := address.FromConfidential(addr)
				script, blindKey = res.Script, res.BlindingKey
			} else {
				script, _ = address.ToOutputScript(addr)
			}
		} else {
			script, _ = hex.DecodeString(out.GetScript())
			blindKey, _ = hex.DecodeString(out.GetBlindingPubkey())
		}
		output := application.Output{
			Asset:       out.GetAsset(),
			Amount:      out.GetAmount(),
			Script:      script,
			BlindingKey: blindKey,
		}
		if err := output.Validate(); err != nil {
			return nil, err
		}
		outputs = append(outputs, output)
	}
	return outputs, nil
}

func parseAmount(amount uint64) (uint64, error) {
	if amount == 0 {
		return 0, fmt.Errorf("missing amount")
	}
	return amount, nil
}

func parseAsset(asset string) (string, error) {
	if len(asset) == 0 {
		return "", fmt.Errorf("missing asset")
	}
	buf, err := hex.DecodeString(asset)
	if err != nil {
		return "", fmt.Errorf("invalid asset format")
	}
	if len(buf) != 32 {
		return "", fmt.Errorf("invalid asset length")
	}
	return asset, nil
}

func parseCoinSelectionStrategy(str pb.SelectUtxosRequest_Strategy) int {
	return application.CoinSelectionStrategySmallestSubset
}

func parseMillisatsPerByte(ratio uint64) (uint64, error) {
	if ratio == 0 {
		return application.MinMillisatsPerByte, nil
	}
	if ratio < application.MinMillisatsPerByte {
		return 0, fmt.Errorf("mSats/byte ratio is too low")
	}
	return ratio, nil
}

func parseTxHex(txHex string) (string, error) {
	if len(txHex) == 0 {
		return "", fmt.Errorf("missing tx hex")
	}
	return txHex, nil
}

func parsePset(ptx string) (string, error) {
	if len(ptx) == 0 {
		return "", fmt.Errorf("missing pset")
	}
	return ptx, nil
}

func parseTxEventType(eventType domain.TransactionEventType) pb.TxEventType {
	switch eventType {
	case domain.TransactionAdded:
		return pb.TxEventType_TX_EVENT_TYPE_BROADCASTED
	case domain.TransactionConfirmed:
		return pb.TxEventType_TX_EVENT_TYPE_CONFIRMED
	case domain.TransactionUnconfirmed:
		return pb.TxEventType_TX_EVENT_TYPE_UNCONFIRMED
	default:
		return pb.TxEventType_TX_EVENT_TYPE_UNSPECIFIED
	}
}

func parseUtxoEventType(eventType domain.UtxoEventType) pb.UtxoEventType {
	switch eventType {
	case domain.UtxoAdded:
		return pb.UtxoEventType_UTXO_EVENT_TYPE_NEW
	case domain.UtxoConfirmed:
		return pb.UtxoEventType_UTXO_EVENT_TYPE_CONFIRMED
	case domain.UtxoLocked:
		return pb.UtxoEventType_UTXO_EVENT_TYPE_LOCKED
	case domain.UtxoUnlocked:
		return pb.UtxoEventType_UTXO_EVENT_TYPE_UNLOCKED
	case domain.UtxoSpent:
		return pb.UtxoEventType_UTXO_EVENT_TYPE_SPENT
	default:
		return pb.UtxoEventType_UTXO_EVENT_TYPE_UNSPECIFIED
	}
}

func parseBlockHeight(height uint32) (uint32, error) {
	if int(height) < 0 {
		return 0, fmt.Errorf("invalid block height")
	}
	return height, nil
}

func parseUnblindedInputs(
	list []*pb.UnblindedInput,
) ([]application.UnblindedInput, error) {
	ins := make([]application.UnblindedInput, 0, len(list))
	for _, l := range list {
		if l.GetAsset() == "" {
			return nil, fmt.Errorf("missing unblinded input asset")
		}
		if _, err := parseAsset(l.GetAsset()); err != nil {
			return nil, fmt.Errorf("invalid unblinded input asset")
		}
		if l.GetAmountBlinder() == "" {
			return nil, fmt.Errorf("missing unblinded input amount blinder")
		}
		if _, err := parseAsset(l.GetAmountBlinder()); err != nil {
			return nil, fmt.Errorf("invalid unblinded input amount blinder")
		}
		if l.GetAssetBlinder() == "" {
			return nil, fmt.Errorf("missing unblinded input asset blinder")
		}
		if _, err := parseAsset(l.GetAssetBlinder()); err != nil {
			return nil, fmt.Errorf("invalid unblinded input asset blinder")
		}
		ins = append(ins, application.UnblindedInput{
			Index:         l.GetIndex(),
			Amount:        l.GetAmount(),
			Asset:         l.GetAsset(),
			AmountBlinder: l.GetAmountBlinder(),
			AssetBlinder:  l.GetAssetBlinder(),
		})
	}
	return ins, nil
}
