package grpc_handler

import (
	"context"
	"fmt"

	"github.com/vulpemventures/go-elements/address"
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
	accountInfo, err := a.appSvc.CreateAccountBIP44(ctx, req.GetLabel(), req.GetUnconfidential())
	if err != nil {
		return nil, err
	}
	masterBlindingKey, _ := accountInfo.GetMasterBlindingKey()
	return &pb.CreateAccountBIP44Response{
		Info: &pb.AccountInfo{
			Namespace:         accountInfo.Namespace,
			Label:             accountInfo.Label,
			Xpubs:             []string{accountInfo.Xpub},
			DerivationPath:    accountInfo.DerivationPath,
			MasterBlindingKey: masterBlindingKey,
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

func (a *account) SetAccountLabel(
	ctx context.Context, req *pb.SetAccountLabelRequest,
) (*pb.SetAccountLabelResponse, error) {
	accountName, err := parseAccountName(req.GetAccountName())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	label, err := parseAccountName(req.GetLabel())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	accountInfo, err := a.appSvc.SetAccountLabel(ctx, accountName, label)
	if err != nil {
		return nil, err
	}
	masterBlindingKey, _ := accountInfo.GetMasterBlindingKey()

	return &pb.SetAccountLabelResponse{
		Info: &pb.AccountInfo{
			Namespace:         accountInfo.Namespace,
			Label:             accountInfo.Label,
			Xpubs:             []string{accountInfo.Xpub},
			DerivationPath:    accountInfo.DerivationPath,
			MasterBlindingKey: masterBlindingKey,
		},
	}, nil
}

func (a *account) SetAccountTemplate(
	ctx context.Context, req *pb.SetAccountTemplateRequest,
) (*pb.SetAccountTemplateResponse, error) {
	return &pb.SetAccountTemplateResponse{}, nil
}

func (a *account) DeriveAddresses(
	ctx context.Context, req *pb.DeriveAddressesRequest,
) (*pb.DeriveAddressesResponse, error) {
	name, err := parseAccountName(req.GetAccountName())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	numOfAddresses := req.GetNumOfAddresses()

	addressesInfo, err := a.appSvc.DeriveAddressesForAccount(
		ctx, name, numOfAddresses,
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
	name, err := parseAccountName(req.GetAccountName())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	numOfAddresses := req.GetNumOfAddresses()

	addressesInfo, err := a.appSvc.DeriveChangeAddressesForAccount(
		ctx, name, numOfAddresses,
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
	addresses := req.GetAddresses()
	scripts := make([][]byte, 0, len(addresses))
	for _, addr := range addresses {
		script, err := address.ToOutputScript(addr)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid address %s", addr))
		}
		scripts = append(scripts, script)
	}

	utxosInfo, err := a.appSvc.ListUtxosForAccount(ctx, name, scripts)
	if err != nil {
		return nil, err
	}
	spendableUtxos := parseUtxos(utxosInfo.Spendable.Info())
	lockedUtxos := parseUtxos(utxosInfo.Locked.Info())
	unconfirmedUtxos := parseUtxos(utxosInfo.Unconfirmed.Info())
	return &pb.ListUtxosResponse{
		SpendableUtxos: &pb.Utxos{
			AccountName: name,
			Utxos:       spendableUtxos,
		},
		LockedUtxos: &pb.Utxos{
			AccountName: name,
			Utxos:       lockedUtxos,
		},
		UnconfirmedUtxos: &pb.Utxos{
			AccountName: name,
			Utxos:       unconfirmedUtxos,
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
