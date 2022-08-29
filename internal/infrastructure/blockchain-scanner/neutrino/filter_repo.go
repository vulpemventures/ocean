package neutrino_scanner

import (
	"context"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/vulpemventures/neutrino-elements/pkg/repository"
	"github.com/xujiajun/nutsdb"
)

const filterBucket = "filters"

type filtersRepo struct {
	store *nutsdb.DB
}

func NewFiltersRepo(store *nutsdb.DB) repository.FilterRepository {
	return newFiltersRepo(store)
}

func newFiltersRepo(store *nutsdb.DB) *filtersRepo {
	return &filtersRepo{
		store: store,
	}
}

func (r *filtersRepo) PutFilter(
	_ context.Context, entry *repository.FilterEntry,
) error {
	return r.store.Update(func(tx *nutsdb.Tx) error {
		key := entry.Key
		hashedKey := btcutil.Hash160(append(key.BlockHash, byte(key.FilterType)))
		return tx.Put(filterBucket, hashedKey, entry.NBytes, nutsdb.Persistent)
	})
}

func (r *filtersRepo) GetFilter(
	_ context.Context, key repository.FilterKey,
) (*repository.FilterEntry, error) {
	var nBytes []byte
	hashedKey := btcutil.Hash160(append(key.BlockHash, byte(key.FilterType)))
	if err := r.store.View(func(tx *nutsdb.Tx) error {
		e, err := tx.Get(filterBucket, hashedKey)
		if err != nil {
			if err == nutsdb.ErrKeyNotFound || err == nutsdb.ErrBucketEmpty {
				return repository.ErrFilterNotFound
			}
			return err
		}
		nBytes = e.Value
		return nil
	}); err != nil {
		return nil, err
	}

	return &repository.FilterEntry{
		Key:    key,
		NBytes: nBytes,
	}, nil
}

func (r *filtersRepo) close() error {
	return r.store.Close()
}
