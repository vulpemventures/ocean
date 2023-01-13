package testutil

import (
	"fmt"
	pb "github.com/vulpemventures/ocean/api-spec/protobuf/gen/go/ocean/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	maxMsgRecvSize = grpc.MaxCallRecvMsgSize(1 * 1024 * 1024 * 200)
)

func GetWalletClient(
	address string,
	noTLS bool,
	tlsCertFile string,
) (pb.WalletServiceClient, func(), error) {
	conn, err := getClientConn(address, noTLS, tlsCertFile)
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() { conn.Close() }
	return pb.NewWalletServiceClient(conn), cleanup, nil
}

func GetAccountClient(
	address string,
	noTLS bool,
	tlsCertFile string,
) (pb.AccountServiceClient, func(), error) {
	conn, err := getClientConn(address, noTLS, tlsCertFile)
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() { conn.Close() }
	return pb.NewAccountServiceClient(conn), cleanup, nil
}

func GetTransactionClient(
	address string,
	noTLS bool,
	tlsCertFile string,
) (pb.TransactionServiceClient, func(), error) {
	conn, err := getClientConn(address, noTLS, tlsCertFile)
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() { conn.Close() }
	return pb.NewTransactionServiceClient(conn), cleanup, nil
}

func GetNotificationsClient(
	address string,
	noTLS bool,
	tlsCertFile string,
) (pb.NotificationServiceClient, func(), error) {
	conn, err := getClientConn(address, noTLS, tlsCertFile)
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() { conn.Close() }
	return pb.NewNotificationServiceClient(conn), cleanup, nil
}

func getClientConn(
	address string,
	noTLS bool,
	tlsCertFile string,
) (*grpc.ClientConn, error) {
	opts := []grpc.DialOption{grpc.WithDefaultCallOptions(maxMsgRecvSize)}

	if noTLS {
		opts = append(opts, grpc.WithInsecure())
	} else {
		if tlsCertFile == "" {
			return nil, fmt.Errorf(
				"missing TLS certificate filepath. Try " +
					"'ocean config set tls_cert_path path/to/tls/certificate'",
			)
		}

		tlsCreds, err := credentials.NewClientTLSFromFile(tlsCertFile, "")
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
