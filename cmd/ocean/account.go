package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	pb "github.com/vulpemventures/ocean/api-spec/protobuf/gen/go/ocean/v1"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var (
	accountName, accountLabel string
	numOfAddresses            uint64
	changeAddresses           bool

	accountCreateCmd = &cobra.Command{
		Use:   "create",
		Short: "create new wallet account",
		Long:  "this command lets you create a new wallet account",
		RunE:  accountCreate,
	}
	accountLabelCmd = &cobra.Command{
		Use:   "label",
		Short: "set label for a wallet account",
		Long: "this command lets you set a label for a wallet account " +
			"that you can then use to refer to it",
		RunE: accountSetLabel,
	}
	accountDeriveAddressesCmd = &cobra.Command{
		Use:   "derive",
		Short: "derive new account address",
		Long: "this command lets you derive new addresses for the given wallet " +
			"account",
		RunE: accountDeriveAddresses,
	}
	accountBalanceCmd = &cobra.Command{
		Use:   "balance",
		Short: "get account balance",
		Long: "this command returns info about the balance of the given account " +
			"(confirmed unconfirmed and locked)",
		RunE: accountBalance,
	}
	accountListAddressesCmd = &cobra.Command{
		Use:   "addresses",
		Short: "list account derived addresses",
		Long: "this command returns the list of all receiving address derived " +
			"for the given account so far",
		RunE: accountListAddresses,
	}
	accountListUtxosCmd = &cobra.Command{
		Use:   "utxos",
		Short: "list account utxos",
		Long: "this command returns the list of all utxos owned by the derived " +
			"addresses of the given account",
		RunE: accountListUtxos,
	}
	accountDeleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "delete account",
		Long: "this command lets you delete an existing account. " +
			"The wallet will loose track of every derived address and utxo history",
		RunE: accountDelete,
	}
	accountCmd = &cobra.Command{
		Use:   "account",
		Short: "interact with ocean account interface",
		Long: "this command lets you create new or delete existing wallet " +
			"accounts, derive or list account addresses, and get info like owned " +
			"utxos or account balance",
	}
)

func init() {
	accountCreateCmd.Flags().StringVarP(
		&accountLabel, "label", "l", "", "label for wallet account",
	)

	accountDeriveAddressesCmd.Flags().Uint64VarP(
		&numOfAddresses, "num-addresses", "n", 0, "number of addresses to derive",
	)
	accountDeriveAddressesCmd.Flags().BoolVarP(
		&changeAddresses, "change", "c", false,
		"whether derive change (internal) addresses",
	)

	accountCmd.PersistentFlags().StringVar(
		&accountName, "account-name", "", "account namespace or label",
	)

	accountDeriveAddressesCmd.MarkPersistentFlagRequired("account-name")
	accountBalanceCmd.MarkPersistentFlagRequired("account-name")
	accountListAddressesCmd.MarkPersistentFlagRequired("account-name")
	accountListUtxosCmd.MarkPersistentFlagRequired("account-name")
	accountDeleteCmd.MarkPersistentFlagRequired("account-name")
	accountLabelCmd.MarkPersistentFlagRequired("account-name")

	accountCmd.AddCommand(
		accountCreateCmd, accountDeriveAddressesCmd, accountBalanceCmd,
		accountListAddressesCmd, accountListUtxosCmd, accountDeleteCmd,
		accountLabelCmd,
	)
}

func accountCreate(cmd *cobra.Command, _ []string) error {
	client, cleanup, err := getAccountClient()
	if err != nil {
		return err
	}
	defer cleanup()

	reply, err := client.CreateAccountBIP44(
		context.Background(), &pb.CreateAccountBIP44Request{
			Label: accountLabel,
		},
	)
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

func accountSetLabel(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing label")
	}

	client, cleanup, err := getAccountClient()
	if err != nil {
		return err
	}
	defer cleanup()

	label := args[0]

	reply, err := client.SetAccountLabel(
		context.Background(), &pb.SetAccountLabelRequest{
			AccountName: accountName,
			Label:       label,
		},
	)
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

func accountDeriveAddresses(cmd *cobra.Command, _ []string) error {
	client, cleanup, err := getAccountClient()
	if err != nil {
		return err
	}
	defer cleanup()

	ctx := context.Background()
	var reply protoreflect.ProtoMessage
	if !changeAddresses {
		reply, err = client.DeriveAddresses(ctx, &pb.DeriveAddressesRequest{
			AccountName:    accountName,
			NumOfAddresses: numOfAddresses,
		})
	} else {
		reply, err = client.DeriveChangeAddresses(ctx, &pb.DeriveChangeAddressesRequest{
			AccountName:    accountName,
			NumOfAddresses: numOfAddresses,
		})
	}
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

func accountBalance(cmd *cobra.Command, _ []string) error {
	client, cleanup, err := getAccountClient()
	if err != nil {
		return err
	}
	defer cleanup()

	reply, err := client.Balance(
		context.Background(), &pb.BalanceRequest{
			AccountName: accountName,
		},
	)
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

func accountListAddresses(cmd *cobra.Command, _ []string) error {
	client, cleanup, err := getAccountClient()
	if err != nil {
		return err
	}
	defer cleanup()

	reply, err := client.ListAddresses(
		context.Background(), &pb.ListAddressesRequest{
			AccountName: accountName,
		},
	)
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

func accountListUtxos(cmd *cobra.Command, _ []string) error {
	client, cleanup, err := getAccountClient()
	if err != nil {
		return err
	}
	defer cleanup()

	reply, err := client.ListUtxos(
		context.Background(), &pb.ListUtxosRequest{
			AccountName: accountName,
		},
	)
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

func accountDelete(cmd *cobra.Command, _ []string) error {
	client, cleanup, err := getAccountClient()
	if err != nil {
		return err
	}
	defer cleanup()

	if _, err := client.DeleteAccount(
		context.Background(), &pb.DeleteAccountRequest{
			AccountName: accountName,
		},
	); err != nil {
		printErr(err)
		return nil
	}

	fmt.Println("account deleted")
	return nil
}
