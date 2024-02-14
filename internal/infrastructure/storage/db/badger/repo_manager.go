package dbbadger

import (
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/dgraph-io/badger/v4/options"
	log "github.com/sirupsen/logrus"
	"github.com/timshannon/badgerhold/v4"
	"github.com/vulpemventures/ocean/internal/core/domain"
	"github.com/vulpemventures/ocean/internal/core/ports"
)

// repoManager holds all the badgerhold stores and domain repositories
// implementations in a single data structure.
type repoManager struct {
	utxoRepository   *utxoRepository
	walletRepository *walletRepository
	txRepository     *transactionRepository
	scriptRepository *scriptRepository

	walletEventHandlers *handlerMap
	utxoEventHandlers   *handlerMap
	txEventHandlers     *handlerMap
	scriptEventHandlers *handlerMap
}

// NewRepoManager is the factory for creating a new badger implementation
// of the ports.RepoManager interface.
// It takes care of creating the db files on disk (or in-memory if no baseDbDir
// is provided - to be used only for testing purposes), and opening and closing
// the connection to them.
func NewRepoManager(baseDbDir string, logger badger.Logger) (ports.RepoManager, error) {
	var walletdbDir, utxoDir, txDir, scriptDir string
	if len(baseDbDir) > 0 {
		walletdbDir = filepath.Join(baseDbDir, "wallet")
		utxoDir = filepath.Join(baseDbDir, "utxos")
		txDir = filepath.Join(baseDbDir, "txs")
		scriptDir = filepath.Join(baseDbDir, "scripts")
	}

	walletDb, err := createDb(walletdbDir, logger)
	if err != nil {
		return nil, fmt.Errorf("opening wallet db: %w", err)
	}

	utxoDb, err := createDb(utxoDir, logger)
	if err != nil {
		return nil, fmt.Errorf("opening utxo db: %w", err)
	}

	txDb, err := createDb(txDir, logger)
	if err != nil {
		return nil, fmt.Errorf("opening tx db: %w", err)
	}
	scriptDb, err := createDb(scriptDir, logger)
	if err != nil {
		return nil, fmt.Errorf("opening external scripts db: %w", err)
	}

	utxoRepo := newUtxoRepository(utxoDb)
	walletRepo := newWalletRepository(walletDb)
	txRepo := newTransactionRepository(txDb)
	scriptRepo := newExternalScriptRepository(scriptDb)

	rm := &repoManager{
		utxoRepository:      utxoRepo,
		walletRepository:    walletRepo,
		txRepository:        txRepo,
		scriptRepository:    scriptRepo,
		walletEventHandlers: newHandlerMap(),
		utxoEventHandlers:   newHandlerMap(),
		txEventHandlers:     newHandlerMap(),
		scriptEventHandlers: newHandlerMap(),
	}

	go rm.listenToWalletEvents()
	go rm.listenToUtxoEvents()
	go rm.listenToTxEvents()
	go rm.listenToScriptEvents()

	return rm, nil
}

func (d *repoManager) UtxoRepository() domain.UtxoRepository {
	return d.utxoRepository
}

func (d *repoManager) WalletRepository() domain.WalletRepository {
	return d.walletRepository
}

func (d *repoManager) TransactionRepository() domain.TransactionRepository {
	return d.txRepository
}

func (d *repoManager) ExternalScriptRepository() domain.ExternalScriptRepository {
	return d.scriptRepository
}

func (rm *repoManager) RegisterHandlerForWalletEvent(
	eventType domain.WalletEventType, handler ports.WalletEventHandler,
) {
	rm.walletEventHandlers.set(int(eventType), handler)
}

func (rm *repoManager) RegisterHandlerForUtxoEvent(
	eventType domain.UtxoEventType, handler ports.UtxoEventHandler,
) {
	rm.utxoEventHandlers.set(int(eventType), handler)
}

func (rm *repoManager) RegisterHandlerForTxEvent(
	eventType domain.TransactionEventType, handler ports.TxEventHandler,
) {
	rm.txEventHandlers.set(int(eventType), handler)
}

func (rm *repoManager) RegisterHandlerForExternalScriptEvent(
	eventType domain.ExternalScriptEventType, handler ports.ScriptEventHandler,
) {
	rm.scriptEventHandlers.set(int(eventType), handler)
}

func (d *repoManager) Reset() {
	d.walletRepository.reset()
	d.utxoRepository.reset()
	d.txRepository.reset()
	d.scriptRepository.reset()
}

func (d *repoManager) Close() {
	d.walletRepository.close()
	d.utxoRepository.close()
	d.txRepository.close()
	d.scriptRepository.close()
}

func (rm *repoManager) listenToWalletEvents() {
	for event := range rm.walletRepository.chEvents {
		if handlers, ok := rm.walletEventHandlers.get(int(event.EventType)); ok {
			for i := range handlers {
				handler := handlers[i]
				go handler.(ports.WalletEventHandler)(event)
			}
		}
	}
}

func (rm *repoManager) listenToUtxoEvents() {
	for event := range rm.utxoRepository.chEvents {
		if handlers, ok := rm.utxoEventHandlers.get(int(event.EventType)); ok {
			for i := range handlers {
				handler := handlers[i]
				go handler.(ports.UtxoEventHandler)(event)
			}
		}
	}
}

func (rm *repoManager) listenToTxEvents() {
	for event := range rm.txRepository.chEvents {
		if handlers, ok := rm.txEventHandlers.get(int(event.EventType)); ok {
			for i := range handlers {
				handler := handlers[i]
				go handler.(ports.TxEventHandler)(event)
			}
		}
	}
}

func (rm *repoManager) listenToScriptEvents() {
	for event := range rm.scriptRepository.chEvents {
		if handlers, ok := rm.scriptEventHandlers.get(int(event.EventType)); ok {
			for i := range handlers {
				handler := handlers[i]
				go handler.(ports.ScriptEventHandler)(event)
			}
		}
	}
}

func createDb(dbDir string, logger badger.Logger) (*badgerhold.Store, error) {
	isInMemory := len(dbDir) <= 0

	opts := badger.DefaultOptions(dbDir)
	opts.Logger = logger

	if isInMemory {
		opts.InMemory = true
	} else {
		opts.Compression = options.ZSTD
	}

	db, err := badgerhold.Open(badgerhold.Options{
		Encoder:          badgerhold.DefaultEncode,
		Decoder:          badgerhold.DefaultDecode,
		SequenceBandwith: 100,
		Options:          opts,
	})
	if err != nil {
		return nil, err
	}

	if !isInMemory {
		ticker := time.NewTicker(30 * time.Minute)

		go func() {
			for {
				<-ticker.C
				if err := db.Badger().RunValueLogGC(0.5); err != nil && err != badger.ErrNoRewrite {
					log.Warnf("garbage collector: %s", err)
				}
			}
		}()
	}

	return db, nil
}

// handlerMap is a util type to prevent race conditions when registering
// or retrieving handlers for events.
type handlerMap struct {
	handlersByEventType map[int][]interface{}
	lock                *sync.RWMutex
}

func newHandlerMap() *handlerMap {
	return &handlerMap{
		handlersByEventType: make(map[int][]interface{}),
		lock:                &sync.RWMutex{},
	}
}

func (m *handlerMap) set(key int, val interface{}) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.handlersByEventType[key] = append(m.handlersByEventType[key], val)
}

func (m *handlerMap) get(key int) ([]interface{}, bool) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	val, ok := m.handlersByEventType[key]
	return val, ok
}
