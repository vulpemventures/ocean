package neutrino_scanner

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/vulpemventures/go-elements/transaction"
	"github.com/vulpemventures/neutrino-elements/pkg/blockservice"
	"github.com/vulpemventures/neutrino-elements/pkg/node"
	"github.com/vulpemventures/neutrino-elements/pkg/protocol"
	"github.com/vulpemventures/neutrino-elements/pkg/repository"
	"github.com/vulpemventures/neutrino-elements/pkg/repository/inmemory"
	"github.com/vulpemventures/ocean/internal/core/domain"
	"github.com/vulpemventures/ocean/internal/core/ports"
)

const (
	userAgent = "neutrino-elements:0.1.0-rc.0"
)

type service struct {
	nodeConfig NodeServiceArgs
	nodeSvc    node.NodeService
	blockSvc   blockservice.BlockService
	scanners   map[string]*scannerService

	filtersRepo repository.FilterRepository
	headersRepo repository.BlockHeaderRepository
	lock        *sync.RWMutex
}

type NodeServiceArgs struct {
	Network             string
	FiltersDatadir      string
	BlockHeadersDatadir string
	Peers               []string
}

func (a NodeServiceArgs) validate() error {
	if a.Network == "" {
		return fmt.Errorf("missing network")
	}
	if a.FiltersDatadir == "" {
		return fmt.Errorf("missing filters datadir")
	}
	if a.BlockHeadersDatadir == "" {
		return fmt.Errorf("missing block headers datadir")
	}
	if len(a.Peers) == 0 {
		return fmt.Errorf("list of peers must not be empty")
	}
	return nil
}

func NewNeutrinoScanner(args NodeServiceArgs) (ports.BlockchainScanner, error) {
	if err := args.validate(); err != nil {
		return nil, err
	}

	filtersDb := inmemory.NewFilterInmemory()
	headersDb := inmemory.NewHeaderInmemory()
	nodeSvc, err := node.New(node.NodeConfig{
		Network:        args.Network,
		UserAgent:      userAgent,
		FiltersDB:      filtersDb,
		BlockHeadersDB: headersDb,
	})
	if err != nil {
		return nil, err
	}
	esploraUrl := esploraUrlFromNetwork(args.Network)
	blockSvc := blockservice.NewEsploraBlockService(esploraUrl)
	scanners := make(map[string]*scannerService)
	lock := &sync.RWMutex{}
	return &service{
		args, nodeSvc, blockSvc, scanners, filtersDb, headersDb, lock,
	}, nil
}

func (s *service) Start() {
	s.nodeSvc.Start(s.nodeConfig.Peers[0])
}

func (s *service) Stop() {
	s.nodeSvc.Stop()
	for _, scanner := range s.scanners {
		scanner.stop()
	}
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
	baseUrl := esploraUrlFromNetwork(s.nodeConfig.Network)
	client := &http.Client{}
	utxos := make([]*domain.Utxo, 0, len(utxoKeys))
	for _, key := range utxoKeys {
		url := fmt.Sprintf("%s/tx/%s", baseUrl, key.TxID)
		resp, err := client.Get(url)
		if err != nil {
			return nil, err
		}

		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		tx := esploraTx{}
		json.Unmarshal(body, &tx)
		confirmedStatus := domain.UtxoStatus{
			BlockHeight: tx.Status.BlockHeight,
			BlockTime:   tx.Status.BlockTimestamp,
			BlockHash:   tx.Status.BlockHash,
		}
		utxo := tx.Outputs[key.VOut].toDomain(key, confirmedStatus)
		utxos = append(utxos, utxo)
	}

	return utxos, nil
}

func (s *service) BroadcastTransaction(txHex string) (string, error) {
	tx, err := transaction.NewTxFromHex(txHex)
	if err != nil {
		return "", fmt.Errorf("invalid tx: %s", err)
	}
	if err := s.nodeSvc.SendTransaction(txHex); err != nil {
		return "", err
	}
	return tx.TxHash().String(), nil
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

func (s *service) getOrCreateScanner(
	accountName string, startingBlock uint32,
) *scannerService {
	s.lock.Lock()
	defer s.lock.Unlock()

	if scannerSvc, ok := s.scanners[accountName]; ok {
		return scannerSvc
	}

	genesisHash := genesisBlockHashForNetwork(s.nodeConfig.Network)
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

func esploraUrlFromNetwork(net string) string {
	if net == "nigiri" {
		return "http://localhost:3001"
	}
	if net == "testnet" {
		return "https://blockstream.info/liquidtestnet/api"
	}
	return "https://blockstream.info/liquid/api"
}

func genesisBlockHashForNetwork(net string) *chainhash.Hash {
	magic := protocol.Networks[net]
	genesis := protocol.GetCheckpoints(magic)[0]
	h, _ := chainhash.NewHashFromStr(genesis)
	return h
}

type esploraTx struct {
	Txid     string          `json:"txid"`
	Version  uint32          `json:"version"`
	Locktime uint32          `json:"locktime"`
	Inputs   []esploraTxIn   `json:"vin"`
	Outputs  []esploraTxOut  `json:"vout"`
	Size     uint32          `json:"size"`
	Weight   uint32          `json:"weight"`
	Fee      uint32          `json:"fee"`
	Status   esploraTxStatus `json:"status"`
}

type esploraTxOut struct {
	Asset           string `json:"asset,omitempty"`
	Value           uint64 `json:"value,omitempty"`
	AssetCommitment string `json:"assetcommitment,omitempty"`
	ValueCommitment string `json:"valuecommitment,omitempty"`
	Script          string `json:"scriptpubkey"`
}

func (o esploraTxOut) toDomain(
	key domain.UtxoKey, confirmedStatus domain.UtxoStatus,
) *domain.Utxo {
	script, _ := hex.DecodeString(o.Script)
	valueCommitment, _ := hex.DecodeString(o.ValueCommitment)
	assetCommitment, _ := hex.DecodeString(o.AssetCommitment)
	return &domain.Utxo{
		UtxoKey:         key,
		Value:           o.Value,
		Asset:           o.Asset,
		AssetCommitment: assetCommitment,
		ValueCommitment: valueCommitment,
		Script:          script,
		ConfirmedStatus: confirmedStatus,
	}
}

type esploraTxIn struct {
	Txid     string       `json:"txid"`
	TxIndex  string       `json:"vout"`
	Prevout  esploraTxOut `json:"prevout"`
	Script   string       `json:"scriptsig"`
	Sequence uint32       `json:"sequence"`
	Witness  []string     `json:"witness"`
}

type esploraTxStatus struct {
	Confirmed      bool   `json:"confirmed"`
	BlockHeight    uint64 `json:"block_height"`
	BlockHash      string `json:"block_hash"`
	BlockTimestamp int64  `json:"block_time"`
}
