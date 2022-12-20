package application

import (
	"context"

	"github.com/equitas-foundation/bamp-ocean/internal/core/domain"
	"github.com/equitas-foundation/bamp-ocean/internal/core/ports"
)

// Notification service has the very simple task of making the event channels
// of the used domain.TransactionRepository and domain.UtxoRepository
// accessible by external clients so that they can get real-time updates on the
// status of the internal wallet.
type NotificationService struct {
	repoManager ports.RepoManager
}

func NewNotificationService(
	repoManager ports.RepoManager,
) *NotificationService {
	return &NotificationService{repoManager}
}

func (ns *NotificationService) GetTxChannel(
	ctx context.Context,
) (chan domain.TransactionEvent, error) {
	return ns.repoManager.TransactionRepository().GetEventChannel(), nil
}

func (ns *NotificationService) GetUtxoChannel(
	ctx context.Context,
) (chan domain.UtxoEvent, error) {
	return ns.repoManager.UtxoRepository().GetEventChannel(), nil
}
