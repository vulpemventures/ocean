package dbbadger

import (
	"context"
	"fmt"
	"sync"

	"github.com/dgraph-io/badger/v3"
	log "github.com/sirupsen/logrus"
	"github.com/timshannon/badgerhold/v4"
	"github.com/vulpemventures/ocean/internal/core/domain"
)

type utxoRepository struct {
	store            *badgerhold.Store
	chEvents         chan domain.UtxoEvent
	externalChEvents chan domain.UtxoEvent
	lock             *sync.Mutex

	log func(format string, a ...interface{})
}

func NewUtxoRepository(store *badgerhold.Store) domain.UtxoRepository {
	return newUtxoRepository(store)
}

func newUtxoRepository(store *badgerhold.Store) *utxoRepository {
	chEvents := make(chan domain.UtxoEvent)
	externalChEvents := make(chan domain.UtxoEvent)
	lock := &sync.Mutex{}
	logFn := func(format string, a ...interface{}) {
		format = fmt.Sprintf("utxo repository: %s", format)
		log.Debugf(format, a...)
	}
	return &utxoRepository{store, chEvents, externalChEvents, lock, logFn}
}

func (r *utxoRepository) AddUtxos(
	ctx context.Context, utxos []*domain.Utxo,
) (int, error) {
	return r.addUtxos(ctx, utxos)
}

func (r *utxoRepository) GetUtxosByKey(
	ctx context.Context, utxoKeys []domain.UtxoKey,
) ([]*domain.Utxo, error) {
	utxos := make([]*domain.Utxo, 0, len(utxoKeys))
	for _, key := range utxoKeys {
		query := badgerhold.Where("TxID").Eq(key.TxID).And("VOut").Eq(key.VOut)
		foundUtxos, err := r.findUtxos(ctx, query)
		if err != nil {
			return nil, err
		}
		if len(foundUtxos) > 0 {
			utxos = append(utxos, foundUtxos[0])
		}
	}

	if len(utxos) == 0 {
		return nil, fmt.Errorf("no utxos found with given keys")
	}
	return utxos, nil
}

func (r *utxoRepository) GetAllUtxos(
	ctx context.Context,
) []*domain.Utxo {
	return r.getAllUtxos(ctx)
}

func (r *utxoRepository) GetSpendableUtxos(
	ctx context.Context,
) ([]*domain.Utxo, error) {
	query := badgerhold.Where("Spent").Eq(false).And("Confirmed").Eq(true).
		And("LockTimestamp").Eq(int64(0))

	return r.findUtxos(ctx, query)
}

func (r *utxoRepository) GetAllUtxosForAccount(
	ctx context.Context, accountName string,
) ([]*domain.Utxo, error) {
	query := badgerhold.Where("AccountName").Eq(accountName)

	return r.findUtxos(ctx, query)
}

func (r *utxoRepository) GetSpendableUtxosForAccount(
	ctx context.Context, accountName string,
) ([]*domain.Utxo, error) {
	query := badgerhold.Where("Spent").Eq(false).And("Confirmed").Eq(true).
		And("LockTimestamp").Eq(int64(0)).And("AccountName").Eq(accountName)

	return r.findUtxos(ctx, query)
}

func (r *utxoRepository) GetLockedUtxosForAccount(
	ctx context.Context, accountName string,
) ([]*domain.Utxo, error) {
	query := badgerhold.Where("Spent").Eq(false).And("LockTimestamp").Gt(int64(0)).
		And("AccountName").Eq(accountName)

	return r.findUtxos(ctx, query)
}

func (r *utxoRepository) GetBalanceForAccount(
	ctx context.Context, accountName string,
) (map[string]*domain.Balance, error) {
	utxos, err := r.GetAllUtxosForAccount(ctx, accountName)
	if err != nil {
		return nil, err
	}

	balance := make(map[string]*domain.Balance)
	for _, u := range utxos {
		if u.IsSpent() {
			continue
		}

		if _, ok := balance[u.Asset]; !ok {
			balance[u.Asset] = &domain.Balance{}
		}

		b := balance[u.Asset]
		if u.IsLocked() {
			b.Locked += u.Value
		} else {
			if u.IsConfirmed() {
				b.Confirmed += u.Value
			} else {
				b.Unconfirmed += u.Value
			}
		}
	}
	return balance, nil
}

func (r *utxoRepository) SpendUtxos(
	ctx context.Context, utxoKeys []domain.UtxoKey,
) (int, error) {
	return r.spendUtxos(ctx, utxoKeys)
}

func (r *utxoRepository) ConfirmUtxos(
	ctx context.Context, utxoKeys []domain.UtxoKey,
) (int, error) {
	return r.confirmUtxos(ctx, utxoKeys)
}

func (r *utxoRepository) LockUtxos(
	ctx context.Context, utxoKeys []domain.UtxoKey, timestamp int64,
) (int, error) {
	return r.lockUtxos(ctx, utxoKeys, timestamp)
}

func (r *utxoRepository) UnlockUtxos(
	ctx context.Context, utxoKeys []domain.UtxoKey,
) (int, error) {
	return r.unlockUtxos(ctx, utxoKeys)
}

func (r *utxoRepository) DeleteUtxosForAccount(
	ctx context.Context, accountName string,
) error {
	query := badgerhold.Where("AccountName").Eq(accountName)

	utxos, err := r.findUtxos(ctx, query)
	if err != nil {
		return err
	}

	utxoKeys := make([]domain.UtxoKey, 0, len(utxos))
	for _, u := range utxos {
		utxoKeys = append(utxoKeys, u.Key())
	}
	return r.deleteUtxos(ctx, utxoKeys)
}

func (r *utxoRepository) GetEventChannel() chan domain.UtxoEvent {
	return r.externalChEvents
}

func (r *utxoRepository) addUtxos(
	ctx context.Context, utxos []*domain.Utxo,
) (int, error) {
	count := 0
	utxosInfo := make([]domain.UtxoInfo, 0)
	for _, u := range utxos {
		done, err := r.insertUtxo(ctx, u)
		if err != nil {
			return -1, err
		}
		if done {
			count++
			utxosInfo = append(utxosInfo, u.Info())
		}
	}

	if count > 0 {
		go r.publishEvent(domain.UtxoEvent{
			EventType: domain.UtxoAdded,
			Utxos:     utxosInfo,
		})
	}

	return count, nil
}

func (r *utxoRepository) getAllUtxos(ctx context.Context) []*domain.Utxo {
	utxos, _ := r.findUtxos(ctx, nil)
	return utxos
}

func (r *utxoRepository) spendUtxos(
	ctx context.Context, utxoKeys []domain.UtxoKey,
) (int, error) {
	count := 0
	utxosInfo := make([]domain.UtxoInfo, 0)
	for _, key := range utxoKeys {
		done, info, err := r.spendUtxo(ctx, key)
		if err != nil {
			return -1, err
		}
		if done {
			count++
			utxosInfo = append(utxosInfo, *info)
		}
	}
	if count > 0 {
		go r.publishEvent(domain.UtxoEvent{
			EventType: domain.UtxoSpent,
			Utxos:     utxosInfo,
		})
	}

	return count, nil
}

func (r *utxoRepository) confirmUtxos(
	ctx context.Context, utxoKeys []domain.UtxoKey,
) (int, error) {
	count := 0
	utxosInfo := make([]domain.UtxoInfo, 0)
	for _, key := range utxoKeys {
		done, info, err := r.confirmUtxo(ctx, key)
		if err != nil {
			return -1, err
		}
		if done {
			count++
			utxosInfo = append(utxosInfo, *info)
		}
	}

	if count > 0 {
		go r.publishEvent(domain.UtxoEvent{
			EventType: domain.UtxoConfirmed,
			Utxos:     utxosInfo,
		})
	}

	return count, nil
}

func (r *utxoRepository) lockUtxos(
	ctx context.Context, utxoKeys []domain.UtxoKey, timestamp int64,
) (int, error) {
	count := 0
	utxosInfo := make([]domain.UtxoInfo, 0)
	for _, key := range utxoKeys {
		done, info, err := r.lockUtxo(ctx, key, timestamp)
		if err != nil {
			return -1, err
		}
		if done {
			count++
			utxosInfo = append(utxosInfo, *info)
		}
	}

	if count > 0 {
		go r.publishEvent(domain.UtxoEvent{
			EventType: domain.UtxoLocked,
			Utxos:     utxosInfo,
		})
	}

	return count, nil
}

func (r *utxoRepository) unlockUtxos(
	ctx context.Context, utxoKeys []domain.UtxoKey,
) (int, error) {
	count := 0
	utxosInfo := make([]domain.UtxoInfo, 0)
	for _, key := range utxoKeys {
		done, info, err := r.unlockUtxo(ctx, key)
		if err != nil {
			return -1, err
		}
		if done {
			count++
			utxosInfo = append(utxosInfo, *info)
		}
	}

	if count > 0 {
		go r.publishEvent(domain.UtxoEvent{
			EventType: domain.UtxoUnlocked,
			Utxos:     utxosInfo,
		})
	}

	return count, nil
}

func (r *utxoRepository) spendUtxo(
	ctx context.Context, key domain.UtxoKey,
) (bool, *domain.UtxoInfo, error) {
	query := badgerhold.Where("TxID").Eq(key.TxID).And("VOut").Eq(key.VOut)
	utxos, err := r.findUtxos(ctx, query)
	if err != nil {
		return false, nil, err
	}

	if utxos == nil {
		return false, nil, nil
	}

	utxo := utxos[0]
	if utxo.IsSpent() {
		return false, nil, nil
	}

	utxo.Spend()
	if err := r.updateUtxo(ctx, utxo); err != nil {
		return false, nil, err
	}

	utxoInfo := utxo.Info()
	return true, &utxoInfo, nil
}

func (r *utxoRepository) confirmUtxo(
	ctx context.Context, key domain.UtxoKey,
) (bool, *domain.UtxoInfo, error) {
	query := badgerhold.Where("TxID").Eq(key.TxID).And("VOut").Eq(key.VOut)
	utxos, err := r.findUtxos(ctx, query)
	if err != nil {
		return false, nil, err
	}

	if utxos == nil {
		return false, nil, nil
	}

	utxo := utxos[0]
	if utxo.IsConfirmed() {
		return false, nil, nil
	}

	utxo.Confirm()
	if err := r.updateUtxo(ctx, utxo); err != nil {
		return false, nil, err
	}

	utxoInfo := utxo.Info()
	return true, &utxoInfo, nil
}

func (r *utxoRepository) lockUtxo(
	ctx context.Context, key domain.UtxoKey, timestamp int64,
) (bool, *domain.UtxoInfo, error) {
	query := badgerhold.Where("TxID").Eq(key.TxID).And("VOut").Eq(key.VOut)
	utxos, err := r.findUtxos(ctx, query)
	if err != nil {
		return false, nil, err
	}

	if utxos == nil {
		return false, nil, nil
	}

	utxo := utxos[0]
	if utxo.IsLocked() {
		return false, nil, nil
	}

	utxo.Lock(timestamp)
	if err := r.updateUtxo(ctx, utxo); err != nil {
		return false, nil, err
	}

	utxoInfo := utxo.Info()
	return true, &utxoInfo, nil
}

func (r *utxoRepository) unlockUtxo(
	ctx context.Context, key domain.UtxoKey,
) (bool, *domain.UtxoInfo, error) {
	query := badgerhold.Where("TxID").Eq(key.TxID).And("VOut").Eq(key.VOut)
	utxos, err := r.findUtxos(ctx, query)
	if err != nil {
		return false, nil, err
	}

	if utxos == nil {
		return false, nil, nil
	}

	utxo := utxos[0]
	if !utxo.IsLocked() {
		return false, nil, nil
	}

	utxo.Unlock()
	if err := r.updateUtxo(ctx, utxo); err != nil {
		return false, nil, err
	}

	utxoInfo := utxo.Info()
	return true, &utxoInfo, nil
}

func (r *utxoRepository) findUtxos(
	ctx context.Context, query *badgerhold.Query,
) ([]*domain.Utxo, error) {
	var list []domain.Utxo
	var utxos []*domain.Utxo
	var err error
	if ctx.Value("tx") != nil {
		tx := ctx.Value("tx").(*badger.Txn)
		err = r.store.TxFind(tx, &utxos, query)
	} else {
		err = r.store.Find(&utxos, query)
	}
	if err != nil {
		if err == badgerhold.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}

	for i := range list {
		u := &list[i]
		utxos = append(utxos, u)
	}
	return utxos, nil
}

func (r *utxoRepository) updateUtxo(
	ctx context.Context, utxo *domain.Utxo,
) error {
	key := utxo.Key()
	query := badgerhold.Where("TxID").Eq(key.TxID).And("VOut").Eq(key.VOut)
	if ctx.Value("tx") != nil {
		tx := ctx.Value("tx").(*badger.Txn)
		return r.store.TxUpdateMatching(
			tx, domain.Utxo{}, query, func(record interface{}) error {
				u := record.(*domain.Utxo)
				*u = *utxo
				return nil
			},
		)
	}

	return r.store.UpdateMatching(domain.Utxo{}, query, func(record interface{}) error {
		u := record.(*domain.Utxo)
		*u = *utxo
		return nil
	})
}

func (r *utxoRepository) insertUtxo(
	ctx context.Context, utxo *domain.Utxo,
) (bool, error) {
	var err error
	if ctx.Value("tx") != nil {
		tx := ctx.Value("tx").(*badger.Txn)
		err = r.store.TxInsert(tx, utxo.Key(), *utxo)
	} else {
		err = r.store.Insert(utxo.Key(), *utxo)
	}

	if err != nil {
		if err == badgerhold.ErrKeyExists {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *utxoRepository) deleteUtxos(
	ctx context.Context, keys []domain.UtxoKey,
) error {
	query := &badgerhold.Query{}
	for _, key := range keys {
		qq := badgerhold.Where("TxID").Eq(key.TxID).And("VOut").Eq(key.VOut)
		query = query.Or(qq)
	}

	if ctx.Value("tx") != nil {
		tx := ctx.Value("tx").(*badger.Txn)
		return r.store.TxDeleteMatching(tx, &domain.Utxo{}, query)
	}

	return r.store.DeleteMatching(&domain.Utxo{}, query)
}

func (r *utxoRepository) publishEvent(event domain.UtxoEvent) {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.log("publish event %s", event.EventType)
	r.chEvents <- event

	// send over channel without blocking in case nobody is listening.
	select {
	case r.externalChEvents <- event:
		return
	default:
		return
	}
}

func (r *utxoRepository) close() {
	r.store.Close()
	close(r.chEvents)
	close(r.externalChEvents)
}
