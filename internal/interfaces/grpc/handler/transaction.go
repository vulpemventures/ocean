package grpc_handler

import (
	"context"
	"fmt"

	pb "github.com/vulpemventures/ocean/api-spec/protobuf/gen/go/ocean/v1"
	"github.com/vulpemventures/ocean/internal/core/application"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type transaction struct {
	appSvc *application.TransactionService
}

func NewTransactionHandler(
	appSvc *application.TransactionService,
) pb.TransactionServiceServer {
	return &transaction{appSvc}
}

func (t *transaction) GetTransaction(
	ctx context.Context, req *pb.GetTransactionRequest,
) (*pb.GetTransactionResponse, error) {
	txid := req.GetTxid()
	if err := validateTxid(txid); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	txInfo, err := t.appSvc.GetTransactionInfo(ctx, txid)
	if err != nil {
		return nil, err
	}
	blockDetails := parseBlockDetails(*txInfo)
	return &pb.GetTransactionResponse{
		TxHex:        txInfo.TxHex,
		BlockDetails: blockDetails,
	}, nil
}

func (t *transaction) SelectUtxos(
	ctx context.Context, req *pb.SelectUtxosRequest,
) (*pb.SelectUtxosResponse, error) {
	accountName, err := parseAccountNamespace(req.GetNamespace())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	targetAmount, err := parseAmount(req.GetTargetAmount())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	targetAsset, err := parseAsset(req.GetTargetAsset())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	strategy := parseCoinSelectionStrategy(req.GetStrategy())

	utxos, change, expirationDate, err := t.appSvc.SelectUtxos(
		ctx, accountName, targetAsset, targetAmount, strategy,
	)
	if err != nil {
		return nil, err
	}
	return &pb.SelectUtxosResponse{
		Utxos:          parseUtxos(utxos.Info()),
		Change:         change,
		ExpirationDate: expirationDate,
	}, nil
}

func (t *transaction) EstimateFees(
	ctx context.Context, req *pb.EstimateFeesRequest,
) (*pb.EstimateFeesResponse, error) {
	inputs, err := parseInputs(req.GetInputs())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	outputs, err := parseOutputs(req.GetOutputs())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	millisatsPerByte, err := parseMillisatsPerByte(req.GetMillisatsPerByte())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	feeAmount, err := t.appSvc.EstimateFees(
		ctx, inputs, outputs, millisatsPerByte,
	)
	if err != nil {
		return nil, err
	}

	return &pb.EstimateFeesResponse{FeeAmount: uint64(feeAmount)}, nil
}

func (t *transaction) SignTransaction(
	ctx context.Context, req *pb.SignTransactionRequest,
) (*pb.SignTransactionResponse, error) {
	txHex, err := parseTxHex(req.GetTxHex())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	signedTx, err := t.appSvc.SignTransaction(ctx, txHex, req.GetSighashType())
	if err != nil {
		return nil, err
	}

	return &pb.SignTransactionResponse{TxHex: signedTx}, nil
}

func (t *transaction) BroadcastTransaction(
	ctx context.Context, req *pb.BroadcastTransactionRequest,
) (*pb.BroadcastTransactionResponse, error) {
	txHex, err := parseTxHex(req.GetTxHex())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	txid, err := t.appSvc.BroadcastTransaction(ctx, txHex)
	if err != nil {
		return nil, err
	}

	return &pb.BroadcastTransactionResponse{Txid: txid}, nil
}

func (t *transaction) CreatePset(
	ctx context.Context, req *pb.CreatePsetRequest,
) (*pb.CreatePsetResponse, error) {
	inputs, err := parseInputs(req.GetInputs())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	outputs, err := parseOutputs(req.GetOutputs())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ptx, err := t.appSvc.CreatePset(ctx, inputs, outputs)
	if err != nil {
		return nil, err
	}
	return &pb.CreatePsetResponse{Pset: ptx}, nil
}

func (t *transaction) UpdatePset(
	ctx context.Context, req *pb.UpdatePsetRequest,
) (*pb.UpdatePsetResponse, error) {
	ptx, err := parsePset(req.GetPset())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	inputs, err := parseInputs(req.GetInputs())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	outputs, err := parseOutputs(req.GetOutputs())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	updatedPtx, err := t.appSvc.UpdatePset(ctx, ptx, inputs, outputs)
	if err != nil {
		return nil, err
	}

	return &pb.UpdatePsetResponse{Pset: updatedPtx}, nil
}

func (t *transaction) BlindPset(
	ctx context.Context, req *pb.BlindPsetRequest,
) (*pb.BlindPsetResponse, error) {
	ptx, err := parsePset(req.GetPset())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	extraUnblindedIns, err := parseUnblindedInputs(req.GetExtraUnblindedInputs())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	blindedPtx, err := t.appSvc.BlindPset(
		ctx, ptx, extraUnblindedIns, req.GetLastBlinder(),
	)
	if err != nil {
		return nil, err
	}

	return &pb.BlindPsetResponse{Pset: blindedPtx}, nil
}

func (t *transaction) SignPset(
	ctx context.Context, req *pb.SignPsetRequest,
) (*pb.SignPsetResponse, error) {
	ptx, err := parsePset(req.GetPset())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	signedPtx, err := t.appSvc.SignPset(ctx, ptx, req.GetSighashType())
	if err != nil {
		return nil, err
	}

	return &pb.SignPsetResponse{Pset: signedPtx}, nil
}

func (t *transaction) Mint(
	ctx context.Context, req *pb.MintRequest,
) (*pb.MintResponse, error) {
	return nil, fmt.Errorf("to be implemented")
}

func (t *transaction) Remint(
	ctx context.Context, req *pb.RemintRequest,
) (*pb.RemintResponse, error) {
	return nil, fmt.Errorf("to be implemented")
}

func (t *transaction) Burn(
	ctx context.Context, req *pb.BurnRequest,
) (*pb.BurnResponse, error) {
	return nil, fmt.Errorf("to be implemented")
}

func (t *transaction) Transfer(
	ctx context.Context, req *pb.TransferRequest,
) (*pb.TransferResponse, error) {
	accountName, err := parseAccountNamespace(req.GetNamespace())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	outputs, err := parseOutputs(req.GetReceivers())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	millisatsPerByte, err := parseMillisatsPerByte(req.GetMillisatsPerByte())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	txHex, err := t.appSvc.Transfer(ctx, accountName, outputs, millisatsPerByte)
	if err != nil {
		return nil, err
	}

	return &pb.TransferResponse{TxHex: txHex}, nil
}

func (t *transaction) PegInAddress(
	ctx context.Context, req *pb.PegInAddressRequest,
) (*pb.PegInAddressResponse, error) {
	return nil, fmt.Errorf("to be implemented")
}

func (t *transaction) ClaimPegIn(
	ctx context.Context, req *pb.ClaimPegInRequest,
) (*pb.ClaimPegInResponse, error) {
	return nil, fmt.Errorf("to be implemented")
}

func validateTxid(txid string) error {
	if txid == "" {
		return fmt.Errorf("missing txid")
	}
	if len(txid) != 64 {
		return fmt.Errorf("invalid txid length")
	}
	return nil
}
