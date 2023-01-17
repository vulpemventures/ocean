package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	pb "github.com/vulpemventures/ocean/api-spec/protobuf/gen/go/ocean/v1"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var (
	label           string
	accountName     string
	numOfAddresses  uint64
	changeAddresses bool

	accountCreateCmd = &cobra.Command{
		Use:   "create",
		Short: "create new wallet account",
		Long: "this command lets you create a new wallet account uniquely " +
			"identified by your choosen name",
		RunE: accountCreate,
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
	accountSetLabelCmd = &cobra.Command{
		Use:   "label",
		Short: "set account label",
		Long:  "this command lets you set label of an existing account",
		RunE:  accountSetLabel,
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
	accountDeriveAddressesCmd.Flags().Uint64VarP(
		&numOfAddresses, "num-addresses", "n", 0, "number of addresses to derive",
	)
	accountDeriveAddressesCmd.Flags().BoolVarP(
		&changeAddresses, "change", "c", false,
		"whether derive change (internal) addresses",
	)

	accountCmd.PersistentFlags().StringVar(&label, "label", "", "account label")
	accountCmd.PersistentFlags().StringVar(&accountName, "account-name", "", "account label or namespace")

	accountCmd.AddCommand(
		accountCreateCmd, accountDeriveAddressesCmd, accountBalanceCmd,
		accountListAddressesCmd, accountListUtxosCmd, accountDeleteCmd,
		accountSetLabelCmd,
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
			Label: label,
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
			Name:           accountName,
			NumOfAddresses: numOfAddresses,
		})
	} else {
		reply, err = client.DeriveChangeAddresses(ctx, &pb.DeriveChangeAddressesRequest{
			Name:           accountName,
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
			Name: accountName,
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
			Name: accountName,
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
			Name: accountName,
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
			Name: accountName,
		},
	); err != nil {
		printErr(err)
		return nil
	}

	fmt.Println("account deleted")
	return nil
}

func accountSetLabel(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return errors.New("missing account label")
	}

	if len(args) > 1 {
		return errors.New("too many arguments, please provide label only")
	}

	client, cleanup, err := getAccountClient()
	if err != nil {
		return err
	}
	defer cleanup()

	_, err = client.SetAccountLabel(context.Background(), &pb.SetAccountLabelRequest{
		Name:  accountName,
		Label: args[0],
	})
	if err != nil {
		printErr(err)
		return nil
	}

	return nil
}
