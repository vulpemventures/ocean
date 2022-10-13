package elements_scanner

import (
	"bytes"
	"context"
	"fmt"
	"github.com/btcsuite/btcd/btcutil/gcs/builder"
	"github.com/vulpemventures/neutrino-elements/pkg/scanner"
	"sync"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/vulpemventures/go-elements/elementsutil"
	"github.com/vulpemventures/go-elements/transaction"
	"github.com/vulpemventures/neutrino-elements/pkg/blockservice"
	"github.com/vulpemventures/neutrino-elements/pkg/protocol"
	"github.com/vulpemventures/neutrino-elements/pkg/repository"
	"github.com/vulpemventures/ocean/internal/core/domain"
	"github.com/vulpemventures/ocean/internal/core/ports"
)

type service struct {
	args      ServiceArgs
	rpcClient *rpcClient
	blockSvc  blockservice.BlockService
	scanners  map[string]*scannerService

	filtersRepo repository.FilterRepository
	headersRepo repository.BlockHeaderRepository
	lock        *sync.RWMutex
}

type ServiceArgs struct {
	RpcAddr             string
	Network             string
	FiltersDatadir      string
	BlockHeadersDatadir string
	EsploraUrl          string
}

func (a ServiceArgs) validate() error {
	if a.RpcAddr == "" {
		return fmt.Errorf("missing elements node rpc address to connect to")
	}
	if a.Network == "" {
		return fmt.Errorf("missing network")
	}
	if a.FiltersDatadir == "" {
		return fmt.Errorf("missing filters datadir")
	}
	if a.BlockHeadersDatadir == "" {
		return fmt.Errorf("missing block headers datadir")
	}
	if a.EsploraUrl == "" {
		return fmt.Errorf("missing esplora url")
	}
	return nil
}

func NewElementsScanner(args ServiceArgs) (ports.BlockchainScanner, error) {
	if err := args.validate(); err != nil {
		return nil, err
	}

	rpcClient, err := newRpcClient(args.RpcAddr, 5)
	if err != nil {
		return nil, err
	}
	filtersDb := newFiltersRepo(rpcClient)
	headersDb := newHeadersRepo(rpcClient)

	blockSvc := blockservice.NewEsploraBlockService(args.EsploraUrl)
	scanners := make(map[string]*scannerService)
	lock := &sync.RWMutex{}
	return &service{
		args, rpcClient, blockSvc, scanners, filtersDb, headersDb, lock,
	}, nil
}

func (s *service) Start() {}

func (s *service) Stop() {}

func (s *service) WatchAddressesForAccount(
	accountName string,
	startingBlockHeight uint32,
	addresses []domain.AddressInfo,
) <-chan scanner.Report {
	scannerSvc := s.getOrCreateScanner(accountName, startingBlockHeight)
	scannerSvc.watchAddresses(addresses)

	return scannerSvc.chReport
}

func (s *service) GetUtxoChannel(accountName string) chan []*domain.Utxo {
	scannerSvc := s.getOrCreateScanner(accountName, 0)
	return scannerSvc.chUtxos
}

func (s *service) GetTxChannel(accountName string) chan *domain.Transaction {
	scannerSvc := s.getOrCreateScanner(accountName, 0)
	return scannerSvc.chTxs
}

func (s *service) WatchForAccount(
	accountName string, startingBlock uint32, addressesInfo []domain.AddressInfo,
) {
	scannerSvc := s.getOrCreateScanner(accountName, startingBlock)
	scannerSvc.watchAddresses(addressesInfo)
}

func (s *service) StopWatchForAccount(accountName string) {
	scannerSvc := s.getOrCreateScanner(accountName, 0)
	scannerSvc.stop()
	s.removeScanner(accountName)
}

func (s *service) GetUtxos(utxoKeys []domain.UtxoKey) ([]*domain.Utxo, error) {
	utxos := make([]*domain.Utxo, 0, len(utxoKeys))
	for _, key := range utxoKeys {
		resp, err := s.rpcClient.call("gettransaction", []interface{}{key.TxID})
		if err != nil {
			return nil, err
		}
		m := resp.(map[string]interface{})

		txHex := m["hex"].(string)
		tx, _ := transaction.NewTxFromHex(txHex)

		out := tx.Outputs[key.VOut]
		utxo := &domain.Utxo{
			UtxoKey: key,
			Script:  out.Script,
		}
		if out.IsConfidential() {
			utxo.AssetCommitment = out.Asset
			utxo.ValueCommitment = out.Value
			utxo.Nonce = out.Nonce
			utxo.RangeProof = out.RangeProof
			utxo.SurjectionProof = out.SurjectionProof
		} else {
			utxo.Asset = elementsutil.AssetHashFromBytes(out.Asset)
			utxo.Value, _ = elementsutil.ValueFromBytes(out.Value)
		}
		confirmations := m["confirmations"].(float64)
		if confirmations > 0 {
			blockHeight := uint64(m["blockheight"].(float64))
			blockTimestamp := int64(m["blocktime"].(float64))
			blockHash := m["blockhash"].(string)
			utxo.ConfirmedStatus = domain.UtxoStatus{
				BlockHeight: blockHeight,
				BlockTime:   blockTimestamp,
				BlockHash:   blockHash,
			}
		}
		utxos = append(utxos, utxo)
	}

	return utxos, nil
}

func (s *service) BroadcastTransaction(txHex string) (string, error) {
	if _, err := transaction.NewTxFromHex(txHex); err != nil {
		return "", fmt.Errorf("invalid tx: %s", err)
	}
	resp, err := s.rpcClient.call("sendrawtransaction", []interface{}{txHex})
	if err != nil {
		return "", err
	}
	txid := resp.(string)
	return txid, nil
}

func (s *service) GetLatestBlock() ([]byte, uint32, error) {
	block, err := s.headersRepo.ChainTip(context.Background())
	if err != nil {
		return nil, 0, err
	}
	hash, _ := block.Hash()
	return hash.CloneBytes(), block.Height, nil
}

func (s *service) GetBlockHeight(blockHash []byte) (uint32, error) {
	hash, err := chainhash.NewHash(blockHash)
	if err != nil {
		return 0, err
	}
	block, err := s.headersRepo.GetBlockHeader(context.Background(), *hash)
	if err != nil {
		return 0, err
	}
	return block.Height, nil
}

func (s *service) GetBlockHash(height uint32) ([]byte, error) {
	hash, err := s.headersRepo.GetBlockHashByHeight(context.Background(), height)
	if err != nil {
		return nil, err
	}
	return hash.CloneBytes(), nil
}

func (s *service) SearchTransactionsForOutputScripts(
	outputScripts [][]byte,
	startingBlockHeight uint32,
) ([]transaction.Transaction, error) {
	result := make([]transaction.Transaction, 0)
	nextHeight := startingBlockHeight
	chainTip, err := s.headersRepo.ChainTip(context.Background())
	if err != nil {
		return nil, err
	}

	for nextHeight <= chainTip.Height {
		var blockHash *chainhash.Hash
		if nextHeight == 0 {
			blockHash = genesisBlockHashForNetwork(s.args.Network)
		} else {
			blockHash, err = s.headersRepo.GetBlockHashByHeight(context.Background(), nextHeight)
			if err != nil {
				return nil, err
			}
		}

		matched, err := s.blockFilterMatches(outputScripts, blockHash)
		if err != nil {
			return nil, err
		}

		foundIndexes := make(map[int]bool)
		if matched {
			txs, fi, err := s.extractBlockMatches(blockHash, outputScripts)
			if err != nil {
				return nil, err
			}

			result = append(result, txs...)
			foundIndexes = fi
		}

		if len(foundIndexes) < len(outputScripts) {
			tmpScripts := make([][]byte, 0)
			for i, v := range outputScripts {
				if !foundIndexes[i] {
					tmpScripts = append(tmpScripts, v)
				}
			}
			outputScripts = tmpScripts
		}

		nextHeight++
		chainTip, err = s.headersRepo.ChainTip(context.Background())
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (s *service) getOrCreateScanner(
	accountName string, startingBlock uint32,
) *scannerService {
	s.lock.Lock()
	defer s.lock.Unlock()

	if scannerSvc, ok := s.scanners[accountName]; ok {
		return scannerSvc
	}

	genesisHash := genesisBlockHashForNetwork(s.args.Network)
	scannerSvc := newScannerSvc(
		accountName, startingBlock, s.filtersRepo, s.headersRepo, s.blockSvc,
		genesisHash,
	)
	s.scanners[accountName] = scannerSvc
	return scannerSvc
}

func (s *service) removeScanner(accountName string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	delete(s.scanners, accountName)
}

func genesisBlockHashForNetwork(net string) *chainhash.Hash {
	magic := protocol.Networks[net]
	genesis := protocol.GetCheckpoints(magic)[0]
	h, _ := chainhash.NewHashFromStr(genesis)
	return h
}

func (s *service) blockFilterMatches(items [][]byte, blockHash *chainhash.Hash) (bool, error) {
	filterToFetchKey := repository.FilterKey{
		BlockHash:  blockHash.CloneBytes(),
		FilterType: repository.RegularFilter,
	}

	filter, err := s.filtersRepo.GetFilter(context.Background(), filterToFetchKey)
	if err != nil {
		if err == repository.ErrFilterNotFound {
			return false, nil
		}
		return false, err
	}

	gcsFilter, err := filter.GcsFilter()
	if err != nil {
		return false, err
	}

	key := builder.DeriveKey(blockHash)
	matched, err := gcsFilter.MatchAny(key, items)
	if err != nil {
		return false, err
	}

	return matched, nil
}

func (s *service) extractBlockMatches(
	blockHash *chainhash.Hash,
	outputScripts [][]byte,
) ([]transaction.Transaction, map[int]bool, error) {
	result := make([]transaction.Transaction, 0)
	resultIndexes := make(map[int]bool)
	block, err := s.blockSvc.GetBlock(blockHash)
	if err != nil {
		return nil, nil, err
	}

	for index, v := range outputScripts {
		for _, tx := range block.TransactionsData.Transactions {
			found := false
			for _, txOutput := range tx.Outputs {
				if bytes.Equal(v, txOutput.Script) {
					found = true
					resultIndexes[index] = true
					break
				}
			}
			if found {
				result = append(result, *tx)
			}
		}
	}

	return result, resultIndexes, nil
}
