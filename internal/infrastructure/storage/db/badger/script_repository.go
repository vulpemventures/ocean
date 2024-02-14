package dbbadger

import (
	"context"
	"fmt"
	"sync"

	"github.com/dgraph-io/badger/v4"
	log "github.com/sirupsen/logrus"
	"github.com/timshannon/badgerhold/v4"
	"github.com/vulpemventures/ocean/internal/core/domain"
)

type scriptRepository struct {
	store    *badgerhold.Store
	chEvents chan domain.ExternalScriptEvent
	lock     *sync.Mutex

	log func(format string, a ...interface{})
}

func NewExternalScriptRepository(
	store *badgerhold.Store,
) domain.ExternalScriptRepository {
	return newExternalScriptRepository(store)
}

func newExternalScriptRepository(
	store *badgerhold.Store,
) *scriptRepository {
	chEvents := make(chan domain.ExternalScriptEvent)
	lock := &sync.Mutex{}
	logFn := func(format string, a ...interface{}) {
		format = fmt.Sprintf("script repository: %s", format)
		log.Debugf(format, a...)
	}
	return &scriptRepository{store, chEvents, lock, logFn}
}

func (r *scriptRepository) AddScript(
	ctx context.Context, info domain.AddressInfo,
) (bool, error) {
	done, err := r.addScript(ctx, info)
	if err != nil {
		return false, err
	}

	if done {
		go r.publishEvent(domain.ExternalScriptEvent{
			EventType: domain.ExternalScriptAdded,
			Info:      info,
		})
	}

	return done, nil
}

func (r *scriptRepository) GetAllScripts(
	ctx context.Context,
) ([]domain.AddressInfo, error) {
	query := &badgerhold.Query{}
	return r.findScripts(ctx, query)
}

func (r *scriptRepository) DeleteScript(
	ctx context.Context, scriptHash string,
) (bool, error) {
	done, err := r.deleteScript(ctx, scriptHash)
	if err != nil {
		return false, err
	}
	if done {
		go r.publishEvent(domain.ExternalScriptEvent{
			EventType: domain.ExternalScriptDeleted,
			Info: domain.AddressInfo{
				Account: scriptHash,
			},
		})
	}

	return done, nil
}

func (r *scriptRepository) addScript(
	ctx context.Context, info domain.AddressInfo,
) (bool, error) {
	var err error
	if ctx.Value("tx") != nil {
		t := ctx.Value("tx").(*badger.Txn)
		err = r.store.TxInsert(t, info.Account, info)
	} else {
		err = r.store.Insert(info.Account, info)
	}

	if err != nil {
		if err == badgerhold.ErrKeyExists {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (r *scriptRepository) findScripts(
	ctx context.Context, query *badgerhold.Query,
) ([]domain.AddressInfo, error) {
	var list []domain.AddressInfo
	var err error
	if ctx.Value("tx") != nil {
		tx := ctx.Value("tx").(*badger.Txn)
		err = r.store.TxFind(tx, &list, query)
	} else {
		err = r.store.Find(&list, query)
	}
	if err != nil {
		if err == badgerhold.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	return list, nil
}

func (r *scriptRepository) deleteScript(
	ctx context.Context, scriptHash string,
) (bool, error) {
	var err error
	if ctx.Value("tx") != nil {
		tx := ctx.Value("tx").(*badger.Txn)
		err = r.store.TxDelete(tx, scriptHash, domain.AddressInfo{})
	} else {
		err = r.store.Delete(scriptHash, domain.AddressInfo{})
	}
	if err != nil {
		if err == badgerhold.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *scriptRepository) publishEvent(event domain.ExternalScriptEvent) {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.log("publish event %s", event.EventType)
	r.chEvents <- event
}

func (r *scriptRepository) reset() {
	r.store.Badger().DropAll()
}

func (r *scriptRepository) close() {
	r.store.Close()
}
