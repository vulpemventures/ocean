package elements_scanner

import (
	"context"
	"fmt"
	"sync"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/equitas-foundation/bamp-ocean/internal/core/domain"
	"github.com/equitas-foundation/bamp-ocean/internal/core/ports"
	"github.com/vulpemventures/go-elements/address"
	"github.com/vulpemventures/go-elements/elementsutil"
	"github.com/vulpemventures/go-elements/network"
	"github.com/vulpemventures/go-elements/transaction"
	"github.com/vulpemventures/neutrino-elements/pkg/blockservice"
	"github.com/vulpemventures/neutrino-elements/pkg/protocol"
	"github.com/vulpemventures/neutrino-elements/pkg/repository"
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

func (a ServiceArgs) network() network.Network {
	switch a.Network {
	case network.Regtest.Name:
		return network.Regtest
	case network.Testnet.Name:
		return network.Testnet
	default:
		return network.Liquid
	}
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

func (s *service) WatchForUtxos(
	accountName string, utxos []domain.UtxoInfo,
) {
	scannerSvc := s.getOrCreateScanner(accountName, 0)
	scannerSvc.watchUtxos(utxos)
}

func (s *service) StopWatchForAccount(accountName string) {
	scannerSvc := s.getOrCreateScanner(accountName, 0)
	scannerSvc.stop()
	s.removeScanner(accountName)
}

func (s *service) GetUtxos(utxoList []domain.Utxo) ([]domain.Utxo, error) {
	utxos := make([]domain.Utxo, 0, len(utxoList))
	for _, u := range utxoList {
		key := u.UtxoKey
		addr := addressFromScript(u.Script, s.args.network())
		if _, err := s.rpcClient.call(
			"importaddress", []interface{}{addr},
		); err != nil {
			return nil, err
		}

		var m map[string]interface{}
		for {
			resp, err := s.rpcClient.call("gettransaction", []interface{}{key.TxID})
			if err != nil {
				continue
			}
			m = resp.(map[string]interface{})
			break
		}

		txHex := m["hex"].(string)
		tx, _ := transaction.NewTxFromHex(txHex)

		out := tx.Outputs[key.VOut]
		utxo := domain.Utxo{
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

func (s *service) GetBlockHash(height uint32) ([]byte, error) {
	hash, err := s.headersRepo.GetBlockHashByHeight(context.Background(), height)
	if err != nil {
		return nil, err
	}
	return hash.CloneBytes(), nil
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

func addressFromScript(script []byte, net network.Network) string {
	switch scriptType := address.GetScriptType(script); scriptType {
	case address.P2PkhScript, address.P2ShScript:
		prefix := net.PubKeyHash
		scriptHash := script[3 : len(script)-2]
		if scriptType == address.P2ShScript {
			prefix = net.ScriptHash
			scriptHash = script[2 : len(script)-1]
		}
		return address.ToBase58(&address.Base58{
			Version: prefix,
			Data:    scriptHash,
		})
	default:
		addr, _ := address.ToBech32(&address.Bech32{
			Prefix:  net.Bech32,
			Version: script[0],
			Program: script[2:],
		})
		return addr
	}
}
