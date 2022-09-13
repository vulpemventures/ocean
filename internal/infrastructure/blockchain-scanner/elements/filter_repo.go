package elements_scanner

import (
	"context"
	"encoding/hex"
	"strings"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/vulpemventures/neutrino-elements/pkg/repository"
)

type filtersRepo struct {
	rpcClient *rpcClient
}

func NewFiltersRepo(rpcClient *rpcClient) repository.FilterRepository {
	return newFiltersRepo(rpcClient)
}

func newFiltersRepo(rpcClient *rpcClient) *filtersRepo {
	return &filtersRepo{rpcClient}
}

func (r *filtersRepo) PutFilter(
	_ context.Context, entry *repository.FilterEntry,
) error {
	return nil
}

func (r *filtersRepo) GetFilter(
	_ context.Context, key repository.FilterKey,
) (*repository.FilterEntry, error) {
	hash, err := chainhash.NewHash(key.BlockHash)
	if err != nil {
		return nil, err
	}
	resp, err := r.rpcClient.call("getblockfilter", []interface{}{hash.String()})
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			return nil, repository.ErrFilterNotFound
		}
		return nil, err
	}

	m := resp.(map[string]interface{})
	filter, ok := m["filter"]
	if !ok {
		return nil, repository.ErrFilterNotFound
	}
	nBytes, _ := hex.DecodeString(filter.(string))
	return &repository.FilterEntry{
		Key:    key,
		NBytes: nBytes,
	}, nil
}
