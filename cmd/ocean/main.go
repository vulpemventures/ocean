package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"

	walletDatadir = btcutil.AppDataDir("ocean-wallet", false)
	initialState  = map[string]string{
		"rpcserver":     "localhost:18000",
		"no_tls":        strconv.FormatBool(false),
		"tls_cert_path": filepath.Join(walletDatadir, "tls", "cert.pem"),
	}

	rootCmd = &cobra.Command{
		Use:   "ocean",
		Short: "CLI for ocean wallet",
		Long:  "This CLI lets you interact with a running ocean wallet daemon",
		PersistentPreRun: func(_ *cobra.Command, _ []string) {
			if _, err := os.Stat(datadir); os.IsNotExist(err) {
				os.Mkdir(datadir, os.ModeDir|0755)
			}
		},
		Version: formatVersion(),
	}
)

func init() {
	rootCmd.AddCommand(configCmd, walletCmd, accountCmd, txCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
