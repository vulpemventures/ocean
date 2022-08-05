package grpc_handler

import (
	"context"
	"fmt"

	pb "github.com/vulpemventures/ocean/api-spec/protobuf/gen/go/ocean/v1"
	"github.com/vulpemventures/ocean/internal/core/application"
)

var ErrStreamConnectionClosed = fmt.Errorf("connection closed on by server")

type notification struct {
	appSvc  *application.NotificationService
	chClose chan struct{}
}

func NewNotificationHandler(
	appSvc *application.NotificationService, chClose chan struct{},
) pb.NotificationServiceServer {
	return &notification{appSvc, chClose}
}

func (n notification) TransactionNotifications(
	req *pb.TransactionNotificationsRequest,
	stream pb.NotificationService_TransactionNotificationsServer,
) error {
	chTxEvents, err := n.appSvc.GetTxChannel(stream.Context())
	if err != nil {
		return err
	}

	for {
		select {
		case e := <-chTxEvents:
			var blockDetails *pb.BlockDetails
			if e.Transaction.IsConfirmed() {
				blockDetails = &pb.BlockDetails{
					Hash:   e.Transaction.BlockHash,
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
		case <-stream.Context().Done():
			return nil
		case <-n.chClose:
			return ErrStreamConnectionClosed
		}
	}
}

func (n notification) UtxosNotifications(
	req *pb.UtxosNotificationsRequest,
	stream pb.NotificationService_UtxosNotificationsServer,
) error {
	chUtxoEvents, err := n.appSvc.GetUtxoChannel(stream.Context())
	if err != nil {
		return err
	}

	for {
		select {
		case e := <-chUtxoEvents:
			if err := stream.Send(&pb.UtxosNotificationsResponse{
				Utxos:     parseUtxos(e.Utxos),
				EventType: parseUtxoEventType(e.EventType),
			}); err != nil {
				return err
			}
		case <-stream.Context().Done():
			return nil
		case <-n.chClose:
			return ErrStreamConnectionClosed
		}
	}
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
