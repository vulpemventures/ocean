package grpc_handler

import (
	"context"

	pb "github.com/vulpemventures/ocean/api-spec/protobuf/gen/go/ocean/v1alpha"
	"github.com/vulpemventures/ocean/internal/core/application"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type account struct {
	appSvc *application.AccountService
}

func NewAccountHandler(appSvc *application.AccountService) pb.AccountServiceServer {
	return &account{
		appSvc: appSvc,
	}
}

func (a *account) CreateAccount(
	ctx context.Context, req *pb.CreateAccountRequest,
) (*pb.CreateAccountResponse, error) {
	name, err := parseAccountName(req.GetName())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	accountInfo, err := a.appSvc.CreateAccount(ctx, name)
	if err != nil {
		return nil, err
	}
	return &pb.CreateAccountResponse{
		AccountName:  accountInfo.Key.Name,
		AccountIndex: accountInfo.Key.Index,
		Xpub:         accountInfo.Xpub,
	}, nil
}

func (a *account) SetAccountTemplate(
	ctx context.Context, req *pb.SetAccountTemplateRequest,
) (*pb.SetAccountTemplateResponse, error) {
	return &pb.SetAccountTemplateResponse{}, nil
}

func (a *account) DeriveAddress(
	ctx context.Context, req *pb.DeriveAddressRequest,
) (*pb.DeriveAddressResponse, error) {
	name, err := parseAccountName(req.GetAccountName())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	numOfAddresses := req.GetNumOfAddresses()

	addressesInfo, err := a.appSvc.DeriveAddressForAccount(
		ctx, name, numOfAddresses,
	)
	if err != nil {
		return nil, err
	}

	return &pb.DeriveAddressResponse{
		Addresses: addressesInfo.Addresses(),
	}, nil
}

func (a *account) DeriveChangeAddress(
	ctx context.Context, req *pb.DeriveChangeAddressRequest,
) (*pb.DeriveChangeAddressResponse, error) {
	name, err := parseAccountName(req.GetAccountName())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	numOfAddresses := req.GetNumOfAddresses()

	addressesInfo, err := a.appSvc.DeriveChangeAddressForAccount(
		ctx, name, numOfAddresses,
	)
	if err != nil {
		return nil, err
	}
	return &pb.DeriveChangeAddressResponse{
		Addresses: addressesInfo.Addresses(),
	}, nil
}

func (a *account) ListAddresses(
	ctx context.Context, req *pb.ListAddressesRequest,
) (*pb.ListAddressesResponse, error) {
	name, err := parseAccountName(req.GetAccountName())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	addressesInfo, err := a.appSvc.ListAddressesForAccount(ctx, name)
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
	name, err := parseAccountName(req.GetAccountName())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	balanceInfo, err := a.appSvc.GetBalanceForAccount(ctx, name)
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
	name, err := parseAccountName(req.GetAccountName())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	utxosInfo, err := a.appSvc.ListUtxosForAccount(ctx, name)
	if err != nil {
		return nil, err
	}
	spendableUtxos := parseUtxos(utxosInfo.Spendable.Info())
	lockedUtxos := parseUtxos(utxosInfo.Locked.Info())
	return &pb.ListUtxosResponse{
		SpendableUtxos: &pb.Utxos{
			AccountName: name,
			Utxos:       spendableUtxos,
		},
		LockedUtxos: &pb.Utxos{
			AccountName: name,
			Utxos:       lockedUtxos,
		},
	}, nil
}

func (a *account) DeleteAccount(
	ctx context.Context, req *pb.DeleteAccountRequest,
) (*pb.DeleteAccountResponse, error) {
	name, err := parseAccountName(req.GetAccountName())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := a.appSvc.DeleteAccount(ctx, name); err != nil {
		return nil, err
	}
	return &pb.DeleteAccountResponse{}, nil
}
