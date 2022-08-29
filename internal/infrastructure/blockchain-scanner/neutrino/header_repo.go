package neutrino_scanner

import (
	"bytes"
	"context"

	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/vulpemventures/go-elements/block"
	"github.com/vulpemventures/neutrino-elements/pkg/repository"
	"github.com/xujiajun/nutsdb"
)

const headersBucket = "headers"

type headersRepo struct {
	store *nutsdb.DB
}

func NewHeadersRepo(store *nutsdb.DB) repository.BlockHeaderRepository {
	return newHeadersRepo(store)
}

func newHeadersRepo(store *nutsdb.DB) *headersRepo {
	return &headersRepo{store}
}

func (r *headersRepo) ChainTip(context.Context) (*block.Header, error) {
	var header *block.Header
	if err := r.store.View(func(tx *nutsdb.Tx) error {
		entries, err := tx.GetAll(headersBucket)
		if err != nil {
			if err == nutsdb.ErrBucketEmpty {
				return repository.ErrNoBlocksHeaders
			}
			return err
		}
		if len(entries) <= 0 {
			return repository.ErrNoBlocksHeaders
		}
		for _, e := range entries {
			h, _ := block.DeserializeHeader(bytes.NewBuffer(e.Value))
			if header == nil {
				header = h
			} else if h.Height > header.Height {
				header = h
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return header, nil
}

func (r *headersRepo) GetBlockHeader(
	_ context.Context, hash chainhash.Hash,
) (*block.Header, error) {
	return r.getHeader(hash)
}

func (r *headersRepo) GetBlockHashByHeight(
	_ context.Context, height uint32,
) (*chainhash.Hash, error) {
	header, err := r.getHeaderByHeight(height)
	if err != nil {
		return nil, err
	}
	hash, err := header.Hash()
	if err != nil {
		return nil, err
	}
	return &hash, nil
}

func (r *headersRepo) WriteHeaders(
	_ context.Context, headers ...block.Header,
) error {
	for _, header := range headers {
		hash, err := header.Hash()
		if err != nil {
			continue
		}
		if err := r.insertHeader(hash, header); err != nil {
			return err
		}
	}
	return nil
}

func (r *headersRepo) LatestBlockLocator(
	ctx context.Context,
) (blockchain.BlockLocator, error) {
	tip, err := r.ChainTip(ctx)
	if err != nil {
		return nil, err
	}
	return r.blockLocatorFromHeader(tip)
}

func (r *headersRepo) HasAllAncestors(_ context.Context, hash chainhash.Hash) (bool, error) {
	header, err := r.getHeader(hash)
	if err != nil {
		return false, err
	}
	if header == nil {
		return false, nil
	}

	for header.Height > 1 {
		prevHash, err := chainhash.NewHash(header.PrevBlockHash)
		if err != nil {
			return false, err
		}
		header, err = r.getHeader(*prevHash)
		if err != nil {
			return false, err
		}
		if header == nil {
			return false, nil
		}
	}
	return true, nil
}

func (r *headersRepo) insertHeader(
	hash chainhash.Hash, header block.Header,
) error {
	buf, _ := header.Serialize()
	return r.store.Update(func(tx *nutsdb.Tx) error {
		return tx.Put(headersBucket, hash[:], buf, nutsdb.Persistent)
	})
}

func (r *headersRepo) getHeaderByHeight(height uint32) (*block.Header, error) {
	var header *block.Header
	if err := r.store.View(func(tx *nutsdb.Tx) error {
		entries, err := tx.GetAll(headersBucket)
		if err != nil {
			return err
		}
		if len(entries) > 0 {
			for _, e := range entries {
				h, _ := block.DeserializeHeader(bytes.NewBuffer(e.Value))
				if h.Height == height {
					header = h
					break
				}
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	if header == nil {
		return nil, repository.ErrBlockNotFound
	}
	return header, nil
}

func (r *headersRepo) getHeader(
	hash chainhash.Hash,
) (*block.Header, error) {
	var header *block.Header
	if err := r.store.View(func(tx *nutsdb.Tx) error {
		e, err := tx.Get(headersBucket, hash[:])
		if err != nil {
			if err == nutsdb.ErrKeyNotFound || err == nutsdb.ErrBucketEmpty {
				return repository.ErrBlockNotFound
			}
			return err
		}
		header, _ = block.DeserializeHeader(bytes.NewBuffer(e.Value))
		return nil
	}); err != nil {
		return nil, err
	}
	return header, nil
}

func (r *headersRepo) blockLocatorFromHeader(
	header *block.Header,
) (blockchain.BlockLocator, error) {
	var locator blockchain.BlockLocator

	hash, err := header.Hash()
	if err != nil {
		return nil, err
	}

	// Append the initial hash
	locator = append(locator, &hash)

	if header.Height == 0 || err != nil {
		return locator, nil
	}

	height := header.Height
	decrement := uint32(1)
	for height > 0 && len(locator) < wire.MaxBlockLocatorsPerMsg {
		blockHeader, err := r.getHeaderByHeight(height)
		if err != nil {
			return nil, err
		}

		headerHash, err := blockHeader.Hash()
		if err != nil {
			return nil, err
		}

		locator = append(locator, &headerHash)

		if decrement > height {
			height = 0
		} else {
			height -= decrement
		}

		// Decrement by 1 for the first 10 blocks, then double the jump
		// until we get to the genesis hash
		if len(locator) > 10 {
			decrement *= 2
		}
	}

	return locator, nil
}

func (r *headersRepo) close() error {
	return r.store.Close()
}
