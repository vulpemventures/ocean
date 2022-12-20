package bamp_cosigner

import (
	"context"

	pb "github.com/equitas-foundation/bamp-ocean/api-spec/protobuf/gen/go/bamp/v1"
	"github.com/equitas-foundation/bamp-ocean/internal/core/ports"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type service struct {
	addr   string
	conn   *grpc.ClientConn
	client pb.CosignerServiceClient
}

func NewService(addr string) (ports.Cosigner, error) {
	conn, err := grpc.Dial(
		addr, grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}
	client := pb.NewCosignerServiceClient(conn)

	svc := &service{addr, conn, client}
	// TODO: call /healthz when available
	return svc, nil
}

func (s *service) GetXpub(ctx context.Context) (string, error) {
	resp, err := s.client.GetXpub(ctx, &pb.GetXpubRequest{})
	if err != nil {
		return "", err
	}
	return resp.GetXpub(), nil
}

func (s *service) RegisterMultiSig(
	ctx context.Context, descriptor string,
) error {
	_, err := s.client.RegisterMultiSig(
		ctx, &pb.RegisterMultiSigRequest{
			WalletDescriptor: descriptor,
		},
	)
	return err
}

func (s *service) SignTx(ctx context.Context, tx string) (string, error) {
	resp, err := s.client.SignTransaction(ctx, &pb.SignTransactionRequest{
		Tx: tx,
	},
	)
	if err != nil {
		return "", err
	}
	return resp.GetSignedTx(), nil
}
