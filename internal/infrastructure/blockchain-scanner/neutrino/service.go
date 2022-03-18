package neutrino_scanner

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
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
	for _, scanner := range s.scanners {
		scanner.stop()
	}
	s.nodeSvc.Stop()
}

func (s *service) GetUtxoChannel(accountName string) chan []*domain.Utxo {
	scannerSvc := s.getOrCreateScanner(accountName)
	return scannerSvc.chUtxos
}

func (s *service) GetTxChannel(accountName string) chan *domain.Transaction {
	scannerSvc := s.getOrCreateScanner(accountName)
	return scannerSvc.chTxs
}

func (s *service) WatchForAccount(
	accountName string, addressesInfo []domain.AddressInfo,
) {
	scannerSvc := s.getOrCreateScanner(accountName)
	scannerSvc.watchAddresses(addressesInfo)
}

func (s *service) StopWatchForAccount(accountName string) {
	scannerSvc := s.getOrCreateScanner(accountName)
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
		utxo := tx.Outputs[key.VOut].toDomain(key, tx.Status.Confirmed)
		utxos = append(utxos, utxo)
	}

	return utxos, nil
}

func (s *service) BroadcastTransaction(txHex string) (string, error) {
	baseUrl := esploraUrlFromNetwork(s.nodeConfig.Network)
	client := &http.Client{}
	url := fmt.Sprintf("%s/tx", baseUrl)
	resp, err := client.Post(url, "text/plain", strings.NewReader(txHex))
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	txid, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(txid), nil
}

func (s *service) getOrCreateScanner(accountName string) *scannerService {
	s.lock.Lock()
	defer s.lock.Unlock()

	if scannerSvc, ok := s.scanners[accountName]; ok {
		return scannerSvc
	}

	genesisHash := genesisBlockHashForNetwork(s.nodeConfig.Network)
	scannerSvc := newScannerSvc(
		accountName, s.filtersRepo, s.headersRepo, s.blockSvc, genesisHash,
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
		return "http://blockstream.info/liquidtestnet/api"
	}
	return "http://blockstream.info/liquid/api"
}

func genesisBlockHashForNetwork(net string) *chainhash.Hash {
	magic := magicFromNetwork(net)
	genesis := protocol.GetCheckpoints(magic)[0]
	h, _ := chainhash.NewHashFromStr(genesis)
	return h
}

func magicFromNetwork(net string) protocol.Magic {
	if net == "nigiri" {
		return protocol.MagicNigiri
	}
	if net == "testnet" {
		return protocol.MagicLiquidTestnet
	}
	return protocol.MagicLiquid
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

func (o esploraTxOut) toDomain(key domain.UtxoKey, confirmed bool) *domain.Utxo {
	script, _ := hex.DecodeString(o.Script)
	valueCommitment, _ := hex.DecodeString(o.ValueCommitment)
	assetCommitment, _ := hex.DecodeString(o.AssetCommitment)
	return &domain.Utxo{
		UtxoKey: domain.UtxoKey{
			TxID: key.TxID,
			VOut: key.VOut,
		},
		Value:           o.Value,
		Asset:           o.Asset,
		AssetCommitment: assetCommitment,
		ValueCommitment: valueCommitment,
		Script:          script,
		Confirmed:       confirmed,
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
	BlockHeight    uint32 `json:"block_height"`
	BlockHash      string `json:"block_hash"`
	BlockTimestamp int64  `json:"block_time"`
}
