package elements_scanner

import (
	"bytes"
	"context"
	"encoding/hex"
	"strings"

	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/vulpemventures/go-elements/block"
	"github.com/vulpemventures/neutrino-elements/pkg/repository"
)

type headersRepo struct {
	rpcClient *rpcClient
}

func NewHeadersRepo(
	rpcClient *rpcClient,
) repository.BlockHeaderRepository {
	return newHeadersRepo(rpcClient)
}

func newHeadersRepo(
	rpcClient *rpcClient,
) *headersRepo {
	return &headersRepo{rpcClient}
}

func (r *headersRepo) ChainTip(
	_ context.Context,
) (*block.Header, error) {
	resp, err := r.rpcClient.call("getbestblockhash", nil)
	if err != nil {
		return nil, err
	}
	hash := resp.(string)

	header, err := r.getHeader(hash)
	if err != nil {
		return nil, err
	}
	if header == nil {
		return nil, repository.ErrNoBlocksHeaders
	}
	return header, nil
}

func (r *headersRepo) GetBlockHeader(
	_ context.Context, hash chainhash.Hash,
) (*block.Header, error) {
	header, err := r.getHeader(hash.String())
	if err != nil {
		return nil, err
	}
	if header == nil {
		return nil, repository.ErrBlockNotFound
	}
	return header, nil
}

func (r *headersRepo) GetBlockHashByHeight(
	_ context.Context, height uint32,
) (*chainhash.Hash, error) {
	return r.getHeaderByHeight(height)
}

func (r *headersRepo) WriteHeaders(
	context.Context, ...block.Header,
) error {
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

func (r *headersRepo) HasAllAncestors(
	_ context.Context, hash chainhash.Hash,
) (bool, error) {
	header, err := r.getHeader(hash.String())
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
		header, err = r.getHeader(prevHash.String())
		if err != nil {
			return false, err
		}
		if header == nil {
			return false, nil
		}
	}
	return true, nil
}

func (r *headersRepo) getHeader(hash string) (*block.Header, error) {
	resp, err := r.rpcClient.call("getblockheader", []interface{}{hash, false})
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			return nil, nil
		}
		return nil, err
	}

	serializedHeader := resp.(string)
	buf, _ := hex.DecodeString(serializedHeader)
	return block.DeserializeHeader(bytes.NewBuffer(buf))
}

func (r *headersRepo) getHeaderByHeight(
	height uint32,
) (*chainhash.Hash, error) {
	resp, err := r.rpcClient.call("getblockhash", []interface{}{height})
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "out of range") {
			return nil, repository.ErrBlockNotFound
		}
		return nil, err
	}

	hash := resp.(string)
	return chainhash.NewHashFromStr(hash)
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
		headerHash, err := r.getHeaderByHeight(height)
		if err != nil {
			return nil, err
		}

		locator = append(locator, headerHash)

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
