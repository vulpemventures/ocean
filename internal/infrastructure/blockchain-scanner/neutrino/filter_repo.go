package neutrino_scanner

import (
	"context"

	"github.com/dgraph-io/badger/v3"
	"github.com/timshannon/badgerhold/v4"
	"github.com/vulpemventures/neutrino-elements/pkg/repository"
)

type filterRepo struct {
	store *badgerhold.Store
}

func NewFilterRepo(store *badgerhold.Store) repository.FilterRepository {
	return newFilterRepo(store)
}

func newFilterRepo(store *badgerhold.Store) *filterRepo {
	return &filterRepo{store}
}

func (r *filterRepo) PutFilter(
	ctx context.Context, entry *repository.FilterEntry,
) error {
	var err error
	if ctx.Value("tx") != nil {
		t := ctx.Value("tx").(*badger.Txn)
		err = r.store.TxInsert(t, entry.Key.String(), *entry)
	} else {
		err = r.store.Insert(entry.Key.String(), *entry)
	}

	if err != nil {
		if err == badgerhold.ErrKeyExists {
			return nil
		}
		return err
	}
	return nil
}

func (r *filterRepo) GetFilter(
	ctx context.Context, key repository.FilterKey,
) (*repository.FilterEntry, error) {
	var err error
	var entry repository.FilterEntry

	if ctx.Value("tx") != nil {
		t := ctx.Value("tx").(*badger.Txn)
		err = r.store.TxGet(t, key.String(), &entry)
	} else {
		err = r.store.Get(key.String(), &entry)
	}

	if err != nil {
		if err == badgerhold.ErrNotFound {
			return nil, repository.ErrFilterNotFound
		}
		return nil, err
	}

	return &entry, nil
}

func (r *filterRepo) close() error {
	return r.store.Close()
}
