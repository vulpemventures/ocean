package grpc_handler

import (
	"context"
	"fmt"

	pb "github.com/vulpemventures/ocean/api-spec/protobuf/gen/go/ocean/v1"
	"github.com/vulpemventures/ocean/internal/core/application"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type account struct {
	appSvc *application.AccountService
}

func NewAccountHandler(appSvc *application.AccountService) pb.AccountServiceServer {
	return &account{appSvc: appSvc}
}

func (a *account) CreateAccountBIP44(
	ctx context.Context, req *pb.CreateAccountBIP44Request,
) (*pb.CreateAccountBIP44Response, error) {
	accountInfo, err := a.appSvc.CreateAccountBIP44(ctx, req.GetLabel())
	if err != nil {
		return nil, err
	}

	return &pb.CreateAccountBIP44Response{
		AccountInfo: &pb.AccountInfo{
			Label:          req.GetLabel(),
			Index:          accountInfo.Key.Index,
			DerivationPath: accountInfo.DerivationPath,
			Xpubs:          []string{accountInfo.Xpub},
			Namespace:      accountInfo.Key.Namespace,
		},
	}, nil
}

func (a *account) CreateAccountMultiSig(
	ctx context.Context, req *pb.CreateAccountMultiSigRequest,
) (*pb.CreateAccountMultiSigResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

func (a *account) CreateAccountCustom(
	ctx context.Context, req *pb.CreateAccountCustomRequest,
) (*pb.CreateAccountCustomResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

func (a *account) SetAccountTemplate(
	ctx context.Context, req *pb.SetAccountTemplateRequest,
) (*pb.SetAccountTemplateResponse, error) {
	return &pb.SetAccountTemplateResponse{}, nil
}

func (a *account) DeriveAddresses(
	ctx context.Context, req *pb.DeriveAddressesRequest,
) (*pb.DeriveAddressesResponse, error) {
	namespace, err := parseAccountNamespace(req.GetNamespace())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	numOfAddresses := req.GetNumOfAddresses()

	addressesInfo, err := a.appSvc.DeriveAddressesForAccount(
		ctx, namespace, numOfAddresses,
	)
	if err != nil {
		return nil, err
	}

	return &pb.DeriveAddressesResponse{
		Addresses: addressesInfo.Addresses(),
	}, nil
}

func (a *account) DeriveChangeAddresses(
	ctx context.Context, req *pb.DeriveChangeAddressesRequest,
) (*pb.DeriveChangeAddressesResponse, error) {
	namespace, err := parseAccountNamespace(req.GetNamespace())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	numOfAddresses := req.GetNumOfAddresses()

	addressesInfo, err := a.appSvc.DeriveChangeAddressesForAccount(
		ctx, namespace, numOfAddresses,
	)
	if err != nil {
		return nil, err
	}
	return &pb.DeriveChangeAddressesResponse{
		Addresses: addressesInfo.Addresses(),
	}, nil
}

func (a *account) ListAddresses(
	ctx context.Context, req *pb.ListAddressesRequest,
) (*pb.ListAddressesResponse, error) {
	namespace, err := parseAccountNamespace(req.GetNamespace())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	addressesInfo, err := a.appSvc.ListAddressesForAccount(ctx, namespace)
	if err != nil {
		return nil, err
	}
	return &pb.ListAddressesResponse{
		Addresses: addressesInfo.Addresses(),
	}, nil
}

func (a *account) Balance(
	ctx context.Context, req *pb.BalanceRequest,
) (*pb.BalanceResponse, error) {
	namespace, err := parseAccountNamespace(req.GetNamespace())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	balanceInfo, err := a.appSvc.GetBalanceForAccount(ctx, namespace)
	if err != nil {
		return nil, err
	}
	balance := make(map[string]*pb.BalanceInfo)
	for asset, b := range balanceInfo {
		balance[asset] = &pb.BalanceInfo{
			ConfirmedBalance:   b.Confirmed,
			UnconfirmedBalance: b.Unconfirmed,
			LockedBalance:      b.Locked,
			TotalBalance:       b.Total(),
		}
	}
	return &pb.BalanceResponse{Balance: balance}, nil
}

func (a *account) ListUtxos(
	ctx context.Context, req *pb.ListUtxosRequest,
) (*pb.ListUtxosResponse, error) {
	namespace, err := parseAccountNamespace(req.GetNamespace())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	utxosInfo, err := a.appSvc.ListUtxosForAccount(ctx, namespace)
	if err != nil {
		return nil, err
	}
	spendableUtxos := parseUtxos(utxosInfo.Spendable.Info())
	lockedUtxos := parseUtxos(utxosInfo.Locked.Info())
	return &pb.ListUtxosResponse{
		SpendableUtxos: &pb.Utxos{
			Namespace: namespace,
			Utxos:     spendableUtxos,
		},
		LockedUtxos: &pb.Utxos{
			Namespace: namespace,
			Utxos:     lockedUtxos,
		},
	}, nil
}

func (a *account) DeleteAccount(
	ctx context.Context, req *pb.DeleteAccountRequest,
) (*pb.DeleteAccountResponse, error) {
	namespace, err := parseAccountNamespace(req.GetNamespace())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := a.appSvc.DeleteAccount(ctx, namespace); err != nil {
		return nil, err
	}
	return &pb.DeleteAccountResponse{}, nil
}

func (a *account) SetAccountLabel(
	ctx context.Context,
	req *pb.SetAccountLabelRequest,
) (*pb.SetAccountLabelResponse, error) {
	namespace, err := parseAccountNamespace(req.GetNamespace())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := a.appSvc.SetAccountLabel(ctx, namespace, req.GetLabel()); err != nil {
		return nil, err
	}

	return &pb.SetAccountLabelResponse{}, nil
}
