package grpc_handler

import (
	"context"

	pb "github.com/vulpemventures/ocean/api-spec/protobuf/gen/go/ocean/v1alpha"
	"github.com/vulpemventures/ocean/internal/core/application"
)

type notification struct {
	appSvc *application.NotificationService
}

func NewNotificationHandler(
	appSvc *application.NotificationService,
) pb.NotificationServiceServer {
	return &notification{appSvc}
}

func (n notification) TransactionNotifications(
	req *pb.TransactionNotificationsRequest,
	stream pb.NotificationService_TransactionNotificationsServer,
) error {
	chTxEvents, err := n.appSvc.GetTxChannel(stream.Context())
	if err != nil {
		return err
	}

	for e := range chTxEvents {
		var blockDetails *pb.BlockDetails
		if e.Transaction.IsConfirmed() {
			blockDetails = &pb.BlockDetails{
				Hash:   []byte(e.Transaction.BlockHash),
				Height: e.Transaction.BlockHeight,
			}
		}
		if err := stream.Send(&pb.TransactionNotificationsResponse{
			AccountNames: e.Transaction.GetAccounts(),
			Txid:         e.Transaction.TxID,
			BlockDetails: blockDetails,
			EventType:    parseTxEventType(e.EventType),
		}); err != nil {
			return err
		}
	}
	return nil
}

func (n notification) UtxosNotifications(
	req *pb.UtxosNotificationsRequest,
	stream pb.NotificationService_UtxosNotificationsServer,
) error {
	chUtxoEvents, err := n.appSvc.GetUtxoChannel(stream.Context())
	if err != nil {
		return err
	}

	for e := range chUtxoEvents {
		if err := stream.Send(&pb.UtxosNotificationsResponse{
			Utxos:     parseUtxos(e.Utxos),
			EventType: parseUtxoEventType(e.EventType),
		}); err != nil {
			return err
		}
	}
	return nil
}

func (n notification) AddWebhook(
	ctx context.Context, req *pb.AddWebhookRequest,
) (*pb.AddWebhookResponse, error) {
	return nil, nil
}

func (n notification) RemoveWebhook(
	ctx context.Context, req *pb.RemoveWebhookRequest,
) (*pb.RemoveWebhookResponse, error) {
	return nil, nil
}

func (n notification) ListWebhooks(
	ctx context.Context, req *pb.ListWebhooksRequest,
) (*pb.ListWebhooksResponse, error) {
	return nil, nil
}
