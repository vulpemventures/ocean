package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math"

	pb "github.com/equitas-foundation/bamp-ocean/api-spec/protobuf/gen/go/ocean/v1"
	"github.com/spf13/cobra"
)

var (
	satsPerByte     float32
	txReceiversJSON []string
	txNoBroadcast   bool

	txTransferCmd = &cobra.Command{
		Use:   "transfer",
		Short: "send funds to other recevivers",
		Long: "this command lets you send your funds to the given receivers " +
			"({asset, amount, addrres})",
		RunE: txTransfer,
	}
	txBroadcastCmd = &cobra.Command{
		Use:   "broadcast",
		Short: "send a transaction over the network to be included in a block",
		Long: "this command lets you publish a final signed transaction " +
			"(in hex format) through the network to be included in a future block",
		RunE: txBroadcast,
	}
	txCmd = &cobra.Command{
		Use:   "transaction",
		Short: "interact with ocean transaction interface",
		Long: "this command lets you send your funds, or use them to mint, " +
			"remint or burn new assets, or even to peg main-chain BTC to the " +
			"Liquid side-chain and unlock LBTC funds",
	}
)

func init() {
	txTransferCmd.Flags().StringArrayVar(
		&txReceiversJSON, "receivers", nil,
		"JSON string list of receivers as "+
			"{\"address\": \"<address>\", \"amount\": <amount in BTC>, \"asset\": \"<asset>\"}",
	)
	txTransferCmd.Flags().BoolVar(&txNoBroadcast, "no-broadcast", false, "use this flag to not broadcast the transaction and get the tx hex instead of its hash")

	txCmd.PersistentFlags().StringVar(
		&accountName, "account-name", "", "name of the account's funds to use",
	)
	txCmd.PersistentFlags().Float32Var(
		&satsPerByte, "sats-per-byte", 0.1, "sats/byte ratio to use for network fees",
	)

	txCmd.AddCommand(txTransferCmd, txBroadcastCmd)
}

func txTransfer(_ *cobra.Command, _ []string) error {
	client, cleanup, err := getTransactionClient()
	if err != nil {
		return err
	}
	defer cleanup()

	ctx := context.Background()
	receivers := make(outputs, 0, len(txReceiversJSON))
	for _, r := range txReceiversJSON {
		receiver := output{}
		if err := json.Unmarshal([]byte(r), &receiver); err != nil {
			printErr(err)
			return nil
		}
		receivers = append(receivers, receiver)
	}

	reply, err := client.Transfer(ctx, &pb.TransferRequest{
		AccountName:      accountName,
		MillisatsPerByte: uint64(satsPerByte * 1000),
		Receivers:        receivers.proto(),
	})
	if err != nil {
		printErr(err)
		return nil
	}

	if txNoBroadcast {
		jsonReply, err := jsonResponse(reply)
		if err != nil {
			printErr(err)
			return nil
		}

		fmt.Println(jsonReply)
		return nil
	}

	bReply, err := client.BroadcastTransaction(
		ctx, &pb.BroadcastTransactionRequest{
			TxHex: reply.GetTxHex(),
		},
	)
	if err != nil {
		printErr(err)
		return nil
	}

	jsonReply, err := jsonResponse(bReply)
	if err != nil {
		printErr(err)
		return nil
	}

	fmt.Println(jsonReply)
	return nil
}

func txBroadcast(_ *cobra.Command, args []string) error {
	if len(args) == 0 {
		printErr(fmt.Errorf("missing tx hex"))
		return nil
	}

	client, cleanup, err := getTransactionClient()
	if err != nil {
		return err
	}
	defer cleanup()

	reply, err := client.BroadcastTransaction(
		context.Background(), &pb.BroadcastTransactionRequest{
			TxHex: args[0],
		})
	if err != nil {
		printErr(err)
		return nil
	}

	jsonReply, err := jsonResponse(reply)
	if err != nil {
		printErr(err)
		return nil
	}

	fmt.Println(jsonReply)
	return nil
}

type output struct {
	Address string  `json:"address"`
	Amount  float64 `json:"amount"`
	Asset   string  `json:"asset"`
}

func (o output) proto() *pb.Output {
	return &pb.Output{
		Address: o.Address,
		Amount:  uint64(o.Amount * math.Pow10(8)),
		Asset:   o.Asset,
	}
}

type outputs []output

func (o outputs) proto() []*pb.Output {
	outs := make([]*pb.Output, 0, len(o))
	for _, out := range o {
		outs = append(outs, out.proto())
	}
	return outs
}
