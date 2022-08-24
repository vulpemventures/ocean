package neutrino_scanner

import (
	"context"

	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/dgraph-io/badger/v3"
	"github.com/timshannon/badgerhold/v4"
	"github.com/vulpemventures/go-elements/block"
	"github.com/vulpemventures/neutrino-elements/pkg/repository"
)

type headersRepo struct {
	store *badgerhold.Store
}

func NewHeadersRepo(store *badgerhold.Store) repository.BlockHeaderRepository {
	return newHeadersRepo(store)
}

func newHeadersRepo(store *badgerhold.Store) *headersRepo {
	return &headersRepo{store}
}

func (r *headersRepo) ChainTip(ctx context.Context) (*block.Header, error) {
	query := &badgerhold.Query{}
	query.SortBy("Height").Reverse().Limit(1)
	header, err := r.findHeader(ctx, query)
	if err != nil {
		return nil, err
	}
	if header == nil {
		return nil, repository.ErrNoBlocksHeaders
	}
	return header, nil
}

func (r *headersRepo) GetBlockHeader(
	ctx context.Context, hash chainhash.Hash,
) (*block.Header, error) {
	header, err := r.getHeader(ctx, hash)
	if err != nil {
		return nil, err
	}
	if header == nil {
		return nil, repository.ErrBlockNotFound
	}
	return header, nil
}

func (r *headersRepo) GetBlockHashByHeight(
	ctx context.Context, height uint32,
) (*chainhash.Hash, error) {
	header, err := r.getHeaderByHeight(ctx, height)
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
	ctx context.Context, headers ...block.Header,
) error {
	for _, header := range headers {
		hash, err := header.Hash()
		if err != nil {
			continue
		}
		if err := r.insertHeader(ctx, hash, header); err != nil {
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
	return r.blockLocatorFromHeader(ctx, tip)
}

func (r *headersRepo) HasAllAncestors(
	ctx context.Context, hash chainhash.Hash,
) (bool, error) {
	header, err := r.getHeader(ctx, hash)
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
		header, err = r.getHeader(ctx, *prevHash)
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
	ctx context.Context, hash chainhash.Hash, header block.Header,
) error {
	var err error
	if ctx.Value("tx") != nil {
		tx := ctx.Value("tx").(*badger.Txn)
		err = r.store.TxInsert(tx, hash.String(), header)
	} else {
		err = r.store.Insert(hash.String(), header)
	}
	if err != nil {
		if err == badgerhold.ErrKeyExists {
			return nil //fmt.Errorf("header with hash %s already exists", hash)
		}
		return err
	}
	return nil
}

func (r *headersRepo) getHeaderByHeight(
	ctx context.Context, height uint32,
) (*block.Header, error) {
	query := badgerhold.Where("Height").Eq(height)
	header, err := r.findHeader(ctx, query)
	if err != nil {
		return nil, err
	}
	if header == nil {
		return nil, repository.ErrBlockNotFound
	}
	return header, nil
}

func (r *headersRepo) getHeader(
	ctx context.Context, hash chainhash.Hash,
) (*block.Header, error) {
	var header block.Header
	var err error
	if ctx.Value("tx") != nil {
		tx := ctx.Value("tx").(*badger.Txn)
		err = r.store.TxGet(tx, hash.String(), &header)
	} else {
		err = r.store.Get(hash.String(), &header)
	}
	if err != nil {
		if err == badgerhold.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &header, nil
}

func (r *headersRepo) findHeader(
	ctx context.Context, query *badgerhold.Query,
) (*block.Header, error) {
	var header []block.Header
	var err error
	if ctx.Value("tx") != nil {
		tx := ctx.Value("tx").(*badger.Txn)
		err = r.store.TxFind(tx, &header, query)
	} else {
		err = r.store.Find(&header, query)
	}
	if err != nil {
		if err == badgerhold.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	if len(header) == 0 {
		return nil, nil
	}

	return &header[0], nil
}

func (r *headersRepo) blockLocatorFromHeader(
	ctx context.Context, header *block.Header,
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
		blockHeader, err := r.getHeaderByHeight(ctx, height)
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
