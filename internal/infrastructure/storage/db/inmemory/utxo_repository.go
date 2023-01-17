package inmemory

import (
	"context"
	"sync"

	"github.com/vulpemventures/ocean/internal/core/domain"
)

type utxoInmemoryStore struct {
	utxosByAccount   map[string][]domain.UtxoKey
	utxos            map[string]*domain.Utxo
	lock             *sync.RWMutex
	walletRepository domain.WalletRepository
}

type utxoRepository struct {
	store            *utxoInmemoryStore
	chEvents         chan domain.UtxoEvent
	externalChEvents chan domain.UtxoEvent
	chLock           *sync.Mutex
}

func NewUtxoRepository(walletRepository domain.WalletRepository) domain.UtxoRepository {
	return newUtxoRepository(walletRepository)
}

func newUtxoRepository(walletRepository domain.WalletRepository) *utxoRepository {
	return &utxoRepository{
		store: &utxoInmemoryStore{
			utxosByAccount:   make(map[string][]domain.UtxoKey),
			utxos:            make(map[string]*domain.Utxo),
			lock:             &sync.RWMutex{},
			walletRepository: walletRepository,
		},
		chEvents:         make(chan domain.UtxoEvent),
		externalChEvents: make(chan domain.UtxoEvent),
		chLock:           &sync.Mutex{},
	}
}

func (r *utxoRepository) AddUtxos(
	_ context.Context, utxos []*domain.Utxo,
) (int, error) {
	r.store.lock.Lock()
	defer r.store.lock.Unlock()

	return r.addUtxos(utxos)
}

func (r *utxoRepository) GetUtxosByKey(
	_ context.Context, utxoKeys []domain.UtxoKey,
) ([]*domain.Utxo, error) {
	r.store.lock.RLock()
	defer r.store.lock.RUnlock()

	utxos := make([]*domain.Utxo, 0, len(utxoKeys))
	for _, key := range utxoKeys {
		u, ok := r.store.utxos[key.Hash()]
		if !ok {
			continue
		}
		utxos = append(utxos, u)
	}

	return utxos, nil
}

func (r *utxoRepository) GetAllUtxos(_ context.Context) ([]*domain.Utxo, error) {
	r.store.lock.RLock()
	defer r.store.lock.RUnlock()

	return r.getUtxos(false), nil
}

func (r *utxoRepository) GetSpendableUtxos(_ context.Context) ([]*domain.Utxo, error) {
	r.store.lock.RLock()
	defer r.store.lock.RUnlock()

	return r.getUtxos(true), nil
}

func (r *utxoRepository) GetAllUtxosForAccount(
	_ context.Context, accountName string,
) ([]*domain.Utxo, error) {
	r.store.lock.RLock()
	defer r.store.lock.RUnlock()

	return r.getUtxosForAccount(accountName, false, false)
}

func (r *utxoRepository) GetSpendableUtxosForAccount(
	_ context.Context, account string,
) ([]*domain.Utxo, error) {
	r.store.lock.RLock()
	defer r.store.lock.RUnlock()

	return r.getUtxosForAccount(account, true, false)
}

func (r *utxoRepository) GetLockedUtxosForAccount(
	_ context.Context, account string,
) ([]*domain.Utxo, error) {
	r.store.lock.RLock()
	defer r.store.lock.RUnlock()

	return r.getUtxosForAccount(account, false, true)
}

func (r *utxoRepository) GetBalanceForAccount(
	_ context.Context, accountName string,
) (map[string]*domain.Balance, error) {
	r.store.lock.RLock()
	defer r.store.lock.RUnlock()

	utxos, _ := r.getUtxosForAccount(accountName, false, false)
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
	_ context.Context, utxos []domain.UtxoKey, status domain.UtxoStatus,
) (int, error) {
	r.store.lock.Lock()
	defer r.store.lock.Unlock()

	return r.spendUtxos(utxos, status)
}

func (r *utxoRepository) ConfirmUtxos(
	_ context.Context, utxos []domain.UtxoKey, status domain.UtxoStatus,
) (int, error) {
	r.store.lock.Lock()
	defer r.store.lock.Unlock()

	return r.confirmUtxos(utxos, status)
}

func (r *utxoRepository) LockUtxos(
	_ context.Context, utxos []domain.UtxoKey, timestamp, expiryTimestamp int64,
) (int, error) {
	r.store.lock.Lock()
	defer r.store.lock.Unlock()

	return r.lockUtxos(utxos, timestamp, expiryTimestamp)
}

func (r *utxoRepository) UnlockUtxos(
	_ context.Context, utxos []domain.UtxoKey,
) (int, error) {
	r.store.lock.Lock()
	defer r.store.lock.Unlock()

	return r.unlockUtxos(utxos)
}

func (r *utxoRepository) DeleteUtxosForAccount(
	_ context.Context, accountName string,
) error {
	account, err := r.getAccount(accountName)
	if err != nil {
		return err
	}

	keys, ok := r.store.utxosByAccount[account]
	if !ok {
		return nil
	}
	for _, key := range keys {
		delete(r.store.utxos, key.Hash())
	}
	delete(r.store.utxosByAccount, account)
	return nil
}

func (r *utxoRepository) GetEventChannel() chan domain.UtxoEvent {
	return r.externalChEvents
}

func (r *utxoRepository) addUtxos(utxos []*domain.Utxo) (int, error) {
	count := 0
	utxosInfo := make([]domain.UtxoInfo, 0, len(utxos))
	for _, u := range utxos {
		if _, ok := r.store.utxos[u.Key().Hash()]; ok {
			continue
		}
		r.store.utxos[u.Key().Hash()] = u
		r.store.utxosByAccount[u.Account] = append(
			r.store.utxosByAccount[u.Account], u.Key(),
		)
		utxosInfo = append(utxosInfo, u.Info())
		count++
	}

	if count > 0 {
		go r.publishEvent(domain.UtxoEvent{
			EventType: domain.UtxoAdded,
			Utxos:     utxosInfo,
		})
	}

	return count, nil
}

func (r *utxoRepository) getUtxos(spendableOnly bool) []*domain.Utxo {
	utxos := make([]*domain.Utxo, 0, len(r.store.utxos))
	for _, u := range r.store.utxos {
		if spendableOnly {
			if !u.IsLocked() && u.IsConfirmed() && !u.IsSpent() {
				utxos = append(utxos, u)
			}
			continue
		}
		utxos = append(utxos, u)
	}
	return utxos
}

func (r *utxoRepository) getUtxosForAccount(
	accountName string, spendableOnly, lockedOnly bool,
) ([]*domain.Utxo, error) {
	account, err := r.getAccount(accountName)
	if err != nil {
		if err == domain.ErrAccountNotFound {
			return nil, nil
		}

		return nil, err
	}

	keys := r.store.utxosByAccount[account]
	if len(keys) == 0 {
		return nil, nil
	}

	utxos := make([]*domain.Utxo, 0, len(keys))
	for _, k := range keys {
		u := r.store.utxos[k.Hash()]

		if spendableOnly {
			if !u.IsLocked() && u.IsConfirmed() && !u.IsSpent() {
				utxos = append(utxos, u)
			}
			continue
		}

		if lockedOnly {
			if u.IsLocked() {
				utxos = append(utxos, u)
			}
			continue
		}
		utxos = append(utxos, u)
	}

	return utxos, nil
}

func (r *utxoRepository) spendUtxos(
	keys []domain.UtxoKey, status domain.UtxoStatus,
) (int, error) {
	count := 0
	utxosInfo := make([]domain.UtxoInfo, 0, len(keys))
	for _, key := range keys {
		utxo, ok := r.store.utxos[key.Hash()]
		if !ok {
			continue
		}

		if utxo.IsSpent() {
			continue
		}

		if err := utxo.Spend(status); err != nil {
			return -1, err
		}

		utxosInfo = append(utxosInfo, utxo.Info())
		count++
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
	keys []domain.UtxoKey, status domain.UtxoStatus,
) (int, error) {
	count := 0
	utxosInfo := make([]domain.UtxoInfo, 0, len(keys))
	for _, key := range keys {
		utxo, ok := r.store.utxos[key.Hash()]
		if !ok {
			continue
		}

		if utxo.IsConfirmed() {
			continue
		}

		if err := utxo.Confirm(status); err != nil {
			return -1, err
		}

		utxosInfo = append(utxosInfo, utxo.Info())
		count++
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
	keys []domain.UtxoKey, timestamp, expiryTimestamp int64,
) (int, error) {
	count := 0
	utxosInfo := make([]domain.UtxoInfo, 0, len(keys))
	for _, key := range keys {
		utxo, ok := r.store.utxos[key.Hash()]
		if !ok {
			continue
		}

		if utxo.IsLocked() {
			continue
		}

		utxo.Lock(timestamp, expiryTimestamp)
		utxosInfo = append(utxosInfo, utxo.Info())
		count++
	}

	if count > 0 {
		go r.publishEvent(domain.UtxoEvent{
			EventType: domain.UtxoLocked,
			Utxos:     utxosInfo,
		})
	}

	return count, nil
}

func (r *utxoRepository) unlockUtxos(keys []domain.UtxoKey) (int, error) {
	count := 0
	utxosInfo := make([]domain.UtxoInfo, 0, len(keys))
	for _, key := range keys {
		utxo, ok := r.store.utxos[key.Hash()]
		if !ok {
			continue
		}

		if !utxo.IsLocked() {
			continue
		}

		utxo.Unlock()
		utxosInfo = append(utxosInfo, utxo.Info())
		count++
	}

	if count > 0 {
		go r.publishEvent(domain.UtxoEvent{
			EventType: domain.UtxoUnlocked,
			Utxos:     utxosInfo,
		})
	}

	return count, nil
}

func (r *utxoRepository) publishEvent(event domain.UtxoEvent) {
	r.chLock.Lock()
	defer r.chLock.Unlock()

	r.chEvents <- event
	// send over channel without blocking in case nobody is listening.
	select {
	case r.externalChEvents <- event:
	default:
	}
}

func (r *utxoRepository) close() {
	close(r.chEvents)
	close(r.externalChEvents)
}

func (r *utxoRepository) getAccount(accountName string) (string, error) {
	wallet, err := r.store.walletRepository.GetWallet(context.Background())
	if err != nil {
		return "", err
	}

	account, err := wallet.GetAccount(accountName)
	if err != nil {
		return "", err
	}

	return account.Info.Namespace, err
}
