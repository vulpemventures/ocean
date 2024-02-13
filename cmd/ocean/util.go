package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"

	pb "github.com/vulpemventures/ocean/api-spec/protobuf/gen/go/ocean/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var (
	maxMsgRecvSize = grpc.MaxCallRecvMsgSize(1 * 1024 * 1024 * 200)
	colorRed       = string("\033[31m")
)

func getWalletClient() (pb.WalletServiceClient, func(), error) {
	conn, err := getClientConn()
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() { conn.Close() }
	return pb.NewWalletServiceClient(conn), cleanup, nil
}

func getAccountClient() (pb.AccountServiceClient, func(), error) {
	conn, err := getClientConn()
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() { conn.Close() }
	return pb.NewAccountServiceClient(conn), cleanup, nil
}

func getTransactionClient() (pb.TransactionServiceClient, func(), error) {
	conn, err := getClientConn()
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() { conn.Close() }
	return pb.NewTransactionServiceClient(conn), cleanup, nil
}

func getClientConn() (*grpc.ClientConn, error) {
	state, err := getState()
	if err != nil {
		return nil, err
	}
	address, ok := state["rpcserver"]
	if !ok {
		return nil, fmt.Errorf("set rpcserver with `config set rpcserver`")
	}

	opts := []grpc.DialOption{grpc.WithDefaultCallOptions(maxMsgRecvSize)}

	noTLS, _ := strconv.ParseBool(state["no_tls"])
	if noTLS {
		opts = append(opts, grpc.WithInsecure())
	} else {
		certPath, ok := state["tls_cert_path"]
		if !ok {
			return nil, fmt.Errorf(
				"missing TLS certificate filepath. Try " +
					"'ocean config set tls_cert_path path/to/tls/certificate'",
			)
		}

		tlsCreds, err := credentials.NewClientTLSFromFile(certPath, "")
		if err != nil {
			return nil, fmt.Errorf("failed to load TLS certificate:  %s", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(tlsCreds))
	}

	conn, err := grpc.Dial(address, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ocean daemon: %v", err)
	}
	return conn, nil
}

func getState() (map[string]string, error) {
	file, err := os.ReadFile(statePath)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		if err := writeState(initialState()); err != nil {
			return nil, err
		}
		return initialState(), nil
	}

	data := map[string]string{}
	json.Unmarshal(file, &data)
	return data, nil
}

func setState(partialState map[string]string) error {
	state, err := getState()
	if err != nil {
		return err
	}

	for key, value := range partialState {
		state[key] = value
	}
	return writeState(state)
}

func writeState(state map[string]string) error {
	dir := filepath.Dir(statePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create directory: %v", err)
		}
	}

	buf, _ := json.MarshalIndent(state, "", "  ")
	if err := os.WriteFile(statePath, buf, 0755); err != nil {
		return fmt.Errorf("writing to file: %w", err)
	}

	return nil
}

func cleanAndExpandPath(path string) string {
	if path == "" {
		return ""
	}

	// Expand initial ~ to OS specific home directory.
	if strings.HasPrefix(path, "~") {
		var homeDir string
		u, err := user.Current()
		if err == nil {
			homeDir = u.HomeDir
		} else {
			homeDir = os.Getenv("HOME")
		}

		path = strings.Replace(path, "~", homeDir, 1)
	}

	// NOTE: The os.ExpandEnv doesn't work with Windows-style %VARIABLE%,
	// but the variables can still be expanded via POSIX-style $VARIABLE.
	return filepath.Clean(os.ExpandEnv(path))
}

func jsonResponse(msg proto.Message) (string, error) {
	buf, err := protojson.MarshalOptions{Multiline: true, EmitUnpopulated: true}.Marshal(msg)
	if err != nil {
		return "", fmt.Errorf("failed to marshal proto message: %s", err)
	}
	return string(buf), nil
}

func printErr(err error) {
	s := status.Convert(err)
	msg := fmt.Sprintf("%s%s", colorRed, capitalize(s.Message()))
	fmt.Fprintln(os.Stderr, msg)
}

func capitalize(s string) string {
	ss := strings.ToUpper(s[0:1])
	ss += s[1:]
	return ss
}

func formatVersion() string {
	return fmt.Sprintf(
		"\nVersion: %s\nCommit: %s\nDate: %s", version, commit, date,
	)
}
