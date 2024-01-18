package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	log "github.com/sirupsen/logrus"
	"github.com/vulpemventures/ocean/internal/core/domain"
	"github.com/vulpemventures/ocean/internal/core/ports"
)

// Notification service has the very simple task of making the event channels
// of the used domain.TransactionRepository and domain.UtxoRepository
// accessible by external clients so that they can get real-time updates on the
// status of the internal wallet.
type NotificationService struct {
	repoManager ports.RepoManager
	bcScanner   ports.BlockchainScanner
	chUtxos     chan domain.UtxoEvent
	chTxs       chan domain.TransactionEvent
	utxoLock    *sync.Mutex
	txLock      *sync.Mutex

	log func(format string, a ...interface{})
}

func NewNotificationService(
	repoManager ports.RepoManager, bcScanner ports.BlockchainScanner,
) *NotificationService {
	logFn := func(format string, a ...interface{}) {
		format = fmt.Sprintf("notification service: %s", format)
		log.Debugf(format, a...)
	}
	chUtxos := make(chan domain.UtxoEvent)
	chTxs := make(chan domain.TransactionEvent)
	utxoLock := &sync.Mutex{}
	txLock := &sync.Mutex{}

	svc := &NotificationService{
		repoManager, bcScanner, chUtxos, chTxs, utxoLock, txLock, logFn,
	}
	svc.registerHandlersForExternalScripts()
	go svc.listenToInternalTxs()
	go svc.listenToInternalUtxos()

	return svc
}

func (ns *NotificationService) GetTxChannel(
	ctx context.Context,
) (chan domain.TransactionEvent, error) {
	return ns.chTxs, nil
}

func (ns *NotificationService) GetUtxoChannel(
	ctx context.Context,
) (chan domain.UtxoEvent, error) {
	return ns.chUtxos, nil
}

func (ns *NotificationService) WatchScript(
	ctx context.Context, scriptHex, blindingKey string,
) (string, error) {
	script, err := hex.DecodeString(scriptHex)
	if err != nil {
		return "", fmt.Errorf("invalid script: must be in hex format")
	}
	var key []byte
	if len(blindingKey) > 0 {
		buf, err := hex.DecodeString(blindingKey)
		if err != nil {
			return "", err
		}
		copy(key, buf)
	}

	buf := sha256.Sum256(script)
	hash, _ := chainhash.NewHash(buf[:])
	label := hash.String()
	externalScript := domain.AddressInfo{
		Account:     label,
		Script:      scriptHex,
		BlindingKey: key,
	}

	done, err := ns.repoManager.ExternalScriptRepository().AddScript(ctx, externalScript)
	if err != nil {
		return "", err
	}
	if done {
		ns.log("added external script with label %s", label)
	}

	return label, nil
}

func (ns *NotificationService) StopWatchingScript(
	ctx context.Context, label string,
) error {
	done, err := ns.repoManager.ExternalScriptRepository().DeleteScript(ctx, label)
	if err != nil {
		return err
	}
	if done {
		ns.log("removed script with label %s", label)
	}
	return nil
}

func (ns *NotificationService) listenToInternalTxs() {
	chTxs := ns.repoManager.TransactionRepository().GetEventChannel()
	for event := range chTxs {
		go ns.publishTx(event)
	}
}

func (ns *NotificationService) listenToInternalUtxos() {
	chUtxos := ns.repoManager.UtxoRepository().GetEventChannel()
	for event := range chUtxos {
		go ns.publishUtxo(event)
	}
}

func (ns *NotificationService) publishUtxo(event domain.UtxoEvent) {
	ns.utxoLock.Lock()
	defer ns.utxoLock.Unlock()

	ns.chUtxos <- event
}

func (ns *NotificationService) publishTx(event domain.TransactionEvent) {
	ns.utxoLock.Lock()
	defer ns.utxoLock.Unlock()

	ns.chTxs <- event
}

func (ns *NotificationService) registerHandlersForExternalScripts() {
	// Start watching external scripts as soon as they are persisted.
	ns.repoManager.RegisterHandlerForExternalScriptEvent(
		domain.ExternalScriptAdded, func(event domain.ExternalScriptEvent) {
			ns.bcScanner.WatchForAccount(event.Info.Account, 0, []domain.AddressInfo{event.Info})
			chUtxos := ns.bcScanner.GetUtxoChannel(event.Info.Account)
			chTxs := ns.bcScanner.GetTxChannel(event.Info.Account)
			go ns.listenToUtxoChannel(event.Info.Account, chUtxos)
			go ns.listenToTxChannel(event.Info.Account, chTxs)
		},
	)

	// Stop watching external scripts as soon as they are removed.
	ns.repoManager.RegisterHandlerForExternalScriptEvent(
		domain.ExternalScriptDeleted, func(event domain.ExternalScriptEvent) {
			ns.bcScanner.StopWatchForAccount(event.Info.Account)
		},
	)

	// Start watching all registered external scripts at startup.
	scripts, _ := ns.repoManager.ExternalScriptRepository().GetAllScripts(
		context.Background(),
	)
	for _, script := range scripts {
		ns.bcScanner.WatchForAccount(script.Account, 0, []domain.AddressInfo{script})
		chUtxos := ns.bcScanner.GetUtxoChannel(script.Account)
		chTxs := ns.bcScanner.GetTxChannel(script.Account)
		go ns.listenToUtxoChannel(script.Account, chUtxos)
		go ns.listenToTxChannel(script.Account, chTxs)
	}
}

func (ns *NotificationService) listenToUtxoChannel(
	scriptHash string, chUtxos chan []*domain.Utxo,
) {
	ns.log("start listening to utxo channel for script %s", scriptHash)

	for utxos := range chUtxos {
		eventType := domain.UtxoAdded
		if utxos[0].IsSpent() {
			eventType = domain.UtxoSpent
		}
		utxoInfo := make([]domain.UtxoInfo, 0, len(utxos))
		for _, u := range utxos {
			utxoInfo = append(utxoInfo, u.Info())
		}
		go ns.publishUtxo(domain.UtxoEvent{
			EventType: eventType,
			Utxos:     utxoInfo,
		})
	}
}

func (ns *NotificationService) listenToTxChannel(
	scriptHash string, chTxs chan *domain.Transaction,
) {
	ns.log("start listening to tx channel for script %s", scriptHash)

	for tx := range chTxs {
		eventType := domain.TransactionConfirmed
		if !tx.IsConfirmed() {
			eventType = domain.TransactionAdded
		}
		go ns.publishTx(domain.TransactionEvent{
			EventType:   eventType,
			Transaction: tx,
		})
	}
}
