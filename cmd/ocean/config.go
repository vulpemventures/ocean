package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var (
	rpcServer   string
	noTLS       bool
	tlsCertPath string

	configSetCmd = &cobra.Command{
		Use:   "set",
		Short: "edit single CLI config entry",
		Long: "this command lets you customize a single configuration entry of " +
			"the ocean wallet CLI",
		Args: cobra.ExactArgs(2),
		RunE: configSet,
	}
	configInitCmd = &cobra.Command{
		Use:   "init",
		Short: "edit multiple CLI config entry",
		Long: "this command lets you customize multiple configuration entres of " +
			"the ocean wallet CLI",
		RunE: configInit,
	}
	configCmd = &cobra.Command{
		Use:   "config",
		Short: "print or edit CLI configuration",
		Long: "this command lets you show or customize the configuration of " +
			"the ocean wallet CLI",
		RunE: configPrint,
	}
)

func init() {
	configInitCmd.Flags().StringVar(
		&rpcServer, "rpcserver", initialState["rpcserver"],
		"address of the ocean wallet to connect to",
	)
	configInitCmd.Flags().BoolVar(
		&noTLS, "no-tls", false,
		"this must be set if the ocean wallet has TLS disabled",
	)
	configInitCmd.Flags().StringVar(
		&tlsCertPath, "tls-cert-path", initialState["tls_cert_path"],
		"the path of the TLS certificate file to use to connect to the ocean "+
			"wallet if it has TLS enabled",
	)
	configCmd.AddCommand(configSetCmd, configInitCmd)
}

func configSet(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := args[1]

	// Prevent setting anything that is not part of the state.
	if _, ok := initialState[key]; !ok {
		return nil
	}

	partialState := map[string]string{key: value}
	if key == "no_tls" {
		partialState["tls_cert_path"] = ""
		if val, _ := strconv.ParseBool(value); !val {
			partialState["tls_cert_path"] = initialState["tls_cert_path"]
		}
	}
	if key == "tls_cert_path" {
		partialState["no_tls"] = "true"
		if len(value) > 0 {
			partialState["no_tls"] = "true"
			value = cleanAndExpandPath(value)
		}
	}
	if err := setState(partialState); err != nil {
		return err
	}

	fmt.Printf("%s %s has been set\n", key, value)

	return nil
}

func configInit(cmd *cobra.Command, args []string) error {
	if _, err := getState(); err != nil {
		return err
	}

	if err := setState(map[string]string{
		"rpcserver":     rpcServer,
		"no_tls":        strconv.FormatBool(noTLS),
		"tls_cert_path": tlsCertPath,
	}); err != nil {
		return err
	}

	fmt.Println("CLI has been configured")

	return nil
}

func configPrint(_ *cobra.Command, _ []string) error {
	state, err := getState()
	if err != nil {
		return err
	}

	buf, _ := json.MarshalIndent(state, "", "   ")
	fmt.Println(string(buf))

	return nil
}
