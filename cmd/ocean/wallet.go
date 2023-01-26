package main

import (
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/spf13/cobra"
	pb "github.com/vulpemventures/ocean/api-spec/protobuf/gen/go/ocean/v1"
)

var (
	datadir   = btcutil.AppDataDir("ocean-cli", false)
	statePath = filepath.Join(datadir, "state.json")

	mnemonic,
	password,
	oldPassword,
	newPassword,
	rootPath string
	birthdayBlock,
	accountThreshold,
	addressThreshold uint32

	walletGenSeedCmd = &cobra.Command{
		Use:   "genseed",
		Short: "generate a random mnemonic",
		Long: "this command lets you generate a new random mnemonic to " +
			"initialize a new wallet from scratch",
		RunE: walletGenSeed,
	}
	walletCreateCmd = &cobra.Command{
		Use:   "create",
		Short: "initialize with a brand new wallet",
		Long: "this command lets you initialize a new ocean wallet from scratch " +
			"with the given mnemonic (or let me create one for you), " +
			"encrypted with your choosen password",
		RunE: walletCreate,
	}
	walletRestoreCmd = &cobra.Command{
		Use:   "restore",
		Short: "restore an existing wallet from a seed",
		Long: "this command lets you restore an ocean wallet from the given " +
			"mnemonic, encrypted with your choosen password",
		RunE: walletRestore,
	}
	walletUnlockCmd = &cobra.Command{
		Use:   "unlock",
		Short: "unlock the wallet",
		Long:  "this command lets you unlock the ocean wallet with your password",
		RunE:  walletUnlock,
	}
	walletLockCmd = &cobra.Command{
		Use:   "lock",
		Short: "lock the wallet",
		Long:  "this command lets you lock the ocean wallet",
		RunE:  walletLock,
	}
	walletChangePwdCmd = &cobra.Command{
		Use:   "changepassword",
		Short: "change the wallet password",
		Long:  "this command lets you change the encryption password of the wallet",
		RunE:  walletChangePwd,
	}
	walletInfoCmd = &cobra.Command{
		Use:   "info",
		Short: "get info about wallet and accounts",
		Long: "this command returns info about the wallet (network, root path " +
			"and master blinding key) and its accounts (key, xpub and derivation path)",
		RunE: walletInfo,
	}
	walletStatusCmd = &cobra.Command{
		Use:   "status",
		Short: "get wallet status",
		Long: "this command returns info about the status of the wallet, like " +
			"if it's initialized, in sync or unlocked",
		RunE: walletStatus,
	}
	authWalletCmd = &cobra.Command{
		Use:   "auth",
		Short: "auth wallet password",
		Long:  "verifies whether the given password is valid without unlocking the wallet",
		RunE:  authWallet,
	}
	walletCmd = &cobra.Command{
		Use:   "wallet",
		Short: "interact with ocean wallet interface",
		Long: "this command lets you initialize, unlock or change the password " +
			"of wallet, as long as retrieving info about its status",
	}
)

func init() {
	walletCreateCmd.Flags().StringVar(
		&mnemonic, "mnemonic", "", "space separated word list as wallet seed",
	)
	walletCreateCmd.Flags().StringVar(&password, "password", "", "encryption password")
	walletCreateCmd.MarkFlagRequired("password")

	walletRestoreCmd.Flags().StringVar(
		&mnemonic, "mnemonic", "", "space separated word list as wallet seed",
	)
	walletRestoreCmd.Flags().StringVar(&password, "password", "", "encryption password")
	walletRestoreCmd.Flags().Uint32Var(
		&birthdayBlock, "birthday-block", 0, "height of the blockchain when wallet was created",
	)
	walletRestoreCmd.Flags().StringVar(&rootPath, "root-path", "", "wallet root path")
	walletRestoreCmd.Flags().Uint32Var(
		&accountThreshold, "account-threshold", 0, "threshold for the number of consecutive accounts to be found empty to consider the restore of the wallet completed",
	)
	walletRestoreCmd.Flags().Uint32Var(
		&addressThreshold, "address-threshold", 0, "threshold for the number of consecutive addresses to be found unused to consider the restore of a wallet account completed",
	)
	walletRestoreCmd.MarkFlagRequired("mnemonic")
	walletRestoreCmd.MarkFlagRequired("password")

	walletUnlockCmd.Flags().StringVar(&password, "password", "", "encryption password")
	walletUnlockCmd.MarkFlagRequired("password")

	walletLockCmd.Flags().StringVar(&password, "password", "", "encryption password")
	walletLockCmd.MarkFlagRequired("password")

	walletChangePwdCmd.Flags().StringVar(&oldPassword, "old-password", "", "current password")
	walletChangePwdCmd.Flags().StringVar(&newPassword, "new-password", "", "new password")
	walletChangePwdCmd.MarkFlagRequired("old-password")
	walletChangePwdCmd.MarkFlagRequired("new-password")

	walletCmd.AddCommand(
		walletGenSeedCmd, walletCreateCmd, walletRestoreCmd, walletUnlockCmd,
		walletLockCmd, walletChangePwdCmd, walletInfoCmd, walletStatusCmd, authWalletCmd,
	)
}

func walletGenSeed(cmd *cobra.Command, args []string) error {
	client, cleanup, err := getWalletClient()
	if err != nil {
		return err
	}
	defer cleanup()

	reply, err := client.GenSeed(context.Background(), &pb.GenSeedRequest{})
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

func walletCreate(cmd *cobra.Command, args []string) error {
	client, cleanup, err := getWalletClient()
	if err != nil {
		return err
	}
	defer cleanup()

	mnemonicToGenerate := len(mnemonic) == 0
	if mnemonicToGenerate {
		reply, err := client.GenSeed(context.Background(), &pb.GenSeedRequest{})
		if err != nil {
			printErr(err)
			return nil
		}
		mnemonic = reply.GetMnemonic()
	}

	if _, err := client.CreateWallet(
		context.Background(), &pb.CreateWalletRequest{
			Mnemonic: mnemonic,
			Password: password,
		},
	); err != nil {
		printErr(err)
		return nil
	}

	if mnemonicToGenerate {
		reply := &pb.GenSeedResponse{Mnemonic: mnemonic}
		jsonReply, err := jsonResponse(reply)
		if err != nil {
			printErr(err)
			return nil
		}

		fmt.Println(jsonReply)
		return nil
	}

	fmt.Println("")
	fmt.Println("wallet initialized")
	return nil
}

func walletRestore(cmd *cobra.Command, args []string) error {
	client, cleanup, err := getWalletClient()
	if err != nil {
		return err
	}
	defer cleanup()

	stream, err := client.RestoreWallet(
		context.Background(), &pb.RestoreWalletRequest{
			Mnemonic:               mnemonic,
			Password:               password,
			BirthdayBlockHeight:    birthdayBlock,
			RootPath:               rootPath,
			EmptyAccountThreshold:  accountThreshold,
			UnusedAddressThreshold: addressThreshold,
		},
	)
	if err != nil {
		printErr(err)
		return nil
	}

	for {
		reply, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		fmt.Println("- message:", reply.GetMessage())
	}

	return nil
}

func walletUnlock(cmd *cobra.Command, args []string) error {
	client, cleanup, err := getWalletClient()
	if err != nil {
		return err
	}
	defer cleanup()

	if _, err := client.Unlock(
		context.Background(), &pb.UnlockRequest{
			Password: password,
		},
	); err != nil {
		printErr(err)
		return nil
	}

	fmt.Println("wallet unlocked")
	return nil
}

func walletLock(cmd *cobra.Command, args []string) error {
	client, cleanup, err := getWalletClient()
	if err != nil {
		return err
	}
	defer cleanup()

	if _, err := client.Lock(
		context.Background(), &pb.LockRequest{
			Password: password,
		},
	); err != nil {
		printErr(err)
		return nil
	}

	fmt.Println("wallet locked")
	return nil
}

func walletChangePwd(cmd *cobra.Command, args []string) error {
	client, cleanup, err := getWalletClient()
	if err != nil {
		return err
	}
	defer cleanup()

	if _, err := client.ChangePassword(
		context.Background(), &pb.ChangePasswordRequest{
			CurrentPassword: oldPassword,
			NewPassword:     newPassword,
		},
	); err != nil {
		printErr(err)
		return nil
	}

	fmt.Println("wallet password updated")
	return nil
}

func walletInfo(cmd *cobra.Command, args []string) error {
	client, cleanup, err := getWalletClient()
	if err != nil {
		return err
	}
	defer cleanup()

	reply, err := client.GetInfo(context.Background(), &pb.GetInfoRequest{})
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

func walletStatus(cmd *cobra.Command, args []string) error {
	client, cleanup, err := getWalletClient()
	if err != nil {
		return err
	}
	defer cleanup()

	reply, err := client.Status(context.Background(), &pb.StatusRequest{})
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

func authWallet(cmd *cobra.Command, _ []string) error {
	client, cleanup, err := getWalletClient()
	if err != nil {
		return err
	}
	defer cleanup()

	reply, err := client.Auth(context.Background(), &pb.AuthRequest{
		Password: password,
	})

	jsonReply, err := jsonResponse(reply)
	if err != nil {
		printErr(err)
		return nil
	}

	fmt.Println(jsonReply)
	return nil
}
