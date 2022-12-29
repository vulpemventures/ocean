package electrum_scanner

import (
	"encoding/hex"
	"fmt"
	"strings"
	"sync"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/equitas-foundation/bamp-ocean/internal/core/domain"
	"github.com/equitas-foundation/bamp-ocean/internal/core/ports"
	log "github.com/sirupsen/logrus"
	"github.com/vulpemventures/go-elements/confidential"
	"github.com/vulpemventures/go-elements/elementsutil"
)

type service struct {
	client electrumClient
	db     *db

	lock *sync.RWMutex

	accountAddressesByScriptHash map[string]map[string]domain.AddressInfo
	utxoChannelByAccount         map[string]chan []*domain.Utxo
	txChannelByAccount           map[string]chan *domain.Transaction
	reportChannelByAccount       map[string]chan accountReport

	log  func(format string, a ...interface{})
	warn func(err error, format string, a ...interface{})
}

type ServiceArgs struct {
	Addr string
}

func (a ServiceArgs) validate() error {
	if a.Addr == "" {
		return fmt.Errorf("missing ws endpoint")
	}
	if !a.withTCP() && !a.withWS() {
		return fmt.Errorf("invalid address: unknown protocol")
	}
	return nil
}

func (a ServiceArgs) withWS() bool {
	return strings.HasPrefix(a.Addr, "ws://") || strings.HasPrefix(a.Addr, "wss://")
}

func (a ServiceArgs) withTCP() bool {
	return strings.HasPrefix(a.Addr, "tcp://") || strings.HasPrefix(a.Addr, "ssl://")
}

func (a ServiceArgs) client() (electrumClient, error) {
	if a.withWS() {
		return newWSClient(a.Addr)
	}
	return newTCPClient(a.Addr)
}

func NewService(args ServiceArgs) (ports.BlockchainScanner, error) {
	if err := args.validate(); err != nil {
		return nil, fmt.Errorf("invalid args: %s", err)
	}

	db := newDb()
	lock := &sync.RWMutex{}
	utxoChannelByAccount := make(map[string]chan []*domain.Utxo)
	txChannelByAccount := make(map[string]chan *domain.Transaction)
	reportChannelByAccount := make(map[string]chan accountReport)
	accountAddressesByScriptHash := make(
		map[string]map[string]domain.AddressInfo,
	)

	client, err := args.client()
	if err != nil {
		return nil, err
	}

	logFn := func(format string, a ...interface{}) {
		format = fmt.Sprintf("scanner: %s", format)
		log.Debugf(format, a...)
	}
	warnFn := func(err error, format string, a ...interface{}) {
		format = fmt.Sprintf("scanner: %s", format)
		log.WithError(err).Warnf(format, a...)
	}

	svc := &service{
		client, db, lock, accountAddressesByScriptHash,
		utxoChannelByAccount, txChannelByAccount, reportChannelByAccount,
		logFn, warnFn,
	}
	svc.db.registerEventHandler(svc.dbEventHandler)

	return svc, nil
}

func (s *service) Start() {
	s.log("start listening to messages from electrum server")

	go s.client.listen()
	s.client.subscribeForBlocks()
}

func (s *service) Stop() {
	s.client.close()
	s.db.close()
	s.log("closed connection with electrum server")
}

func (s *service) WatchForAccount(
	accountName string, _ uint32,
	addresses []domain.AddressInfo,
) {
	accountCh, txHistory := s.client.subscribeForAccount(accountName, addresses)
	if _, ok := s.getAccountChannel(accountName); !ok {
		s.setAccountChannels(accountName, accountCh)

		go s.listenToAccountChannel(accountCh)
	}

	for scriptHash, history := range txHistory {
		s.db.updateAccountTxHistory(accountName, scriptHash, history)
	}
	s.setAddressesByScriptHash(accountName, addresses)
}

func (s *service) WatchForUtxos(
	accountName string, utxos []domain.UtxoInfo,
) {
}

func (s *service) StopWatchForAccount(accountName string) {
	s.client.unsubscribeForAccount(accountName)
}

func (s *service) GetUtxoChannel(accountName string) chan []*domain.Utxo {
	return s.getUtxoChannelByAccount(accountName)
}

func (s *service) GetTxChannel(accountName string) chan *domain.Transaction {
	return s.getTxChannelByAccount(accountName)
}

func (s *service) GetLatestBlock() ([]byte, uint32, error) {
	return s.client.getLatestBlock()
}

// GetBlockHash returns the hash of the block identified by its height.
func (s *service) GetBlockHash(height uint32) ([]byte, error) {
	hash, _, err := s.client.getBlockInfo(height)
	if err != nil {
		return nil, err
	}
	return hash[:], nil
}

// GetUtxos is a sync function to get info about the utxos represented by
// given outpoints (UtxoKeys).
func (s *service) GetUtxos(utxos []domain.Utxo) ([]domain.Utxo, error) {
	return s.client.getUtxos(utxos)
}

// BroadcastTransaction sends the given raw tx (in hex string) over the
// network in order to be included in a later block of the Liquid blockchain.
func (s *service) BroadcastTransaction(txHex string) (string, error) {
	return s.client.broadcastTx(txHex)
}

func (s *service) getAccountChannel(
	account string,
) (chan accountReport, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	ch, ok := s.reportChannelByAccount[account]
	return ch, ok
}

func (s *service) getUtxoChannelByAccount(account string) chan []*domain.Utxo {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.utxoChannelByAccount[account]
}

func (s *service) getTxChannelByAccount(account string) chan *domain.Transaction {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.txChannelByAccount[account]
}

func (s *service) setAccountChannels(
	account string, chReports chan accountReport,
) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.reportChannelByAccount[account] = chReports
	s.utxoChannelByAccount[account] = make(chan []*domain.Utxo)
	s.txChannelByAccount[account] = make(chan *domain.Transaction)
}

func (s *service) listenToAccountChannel(chReports chan accountReport) {
	for report := range chReports {
		history, err := s.client.getScriptHashesHistory(
			[]string{report.scriptHash},
		)
		if err != nil {
			s.warn(
				err, "failed to get history for script hash %s", report.scriptHash,
			)
			continue
		}

		s.db.updateAccountTxHistory(
			report.account, report.scriptHash, history[report.scriptHash],
		)
	}
}

func (s *service) dbEventHandler(event dbEvent) {
	tx, err := s.client.getTx(event.tx.Txid)
	if err != nil {
		s.warn(err, "failed to fetch tx for event %+v", event)
		return
	}

	newUtxos := make([]*domain.Utxo, 0)
	spentUtxos := make([]*domain.Utxo, 0)
	confirmedUtxos := make([]*domain.Utxo, 0)
	var blockHash *chainhash.Hash
	var blockTimestamp int64

	if event.tx.Height > 0 {
		blockHash, blockTimestamp, err = s.client.getBlockInfo(
			uint32(event.tx.Height),
		)
		if err != nil {
			s.warn(err, "failed to fetch block %d", event.tx.Height)
			return
		}

		for _, in := range tx.Inputs {
			spentUtxos = append(spentUtxos, &domain.Utxo{
				UtxoKey: domain.UtxoKey{
					TxID: elementsutil.TxIDFromBytes(in.Hash),
					VOut: in.Index,
				},
				SpentStatus: domain.UtxoStatus{
					Txid:        event.tx.Txid,
					BlockHash:   blockHash.String(),
					BlockTime:   blockTimestamp,
					BlockHeight: uint64(event.tx.Height),
				},
			})
		}
	}

	if event.eventType == txConfirmed {
		for i, out := range tx.Outputs {
			if len(out.Script) > 0 {
				scriptHash := calcScriptHash(hex.EncodeToString(out.Script))
				addrInfo := s.getAddressByScriptHash(event.account, scriptHash)
				if addrInfo != nil {
					confirmedUtxos = append(confirmedUtxos, &domain.Utxo{
						UtxoKey: domain.UtxoKey{
							TxID: event.tx.Txid,
							VOut: uint32(i),
						},
						ConfirmedStatus: domain.UtxoStatus{
							BlockHash:   blockHash.String(),
							BlockTime:   blockTimestamp,
							BlockHeight: uint64(event.tx.Height),
						},
					})
				}
			}
		}
	}

	if event.eventType == txAdded {
		for i, out := range tx.Outputs {
			if len(out.Script) > 0 {
				scriptHash := calcScriptHash(hex.EncodeToString(out.Script))
				addrInfo := s.getAddressByScriptHash(event.account, scriptHash)
				if addrInfo != nil {
					var nonce, valueCommit, assetCommit []byte
					if out.IsConfidential() {
						nonce, valueCommit, assetCommit = out.Nonce, out.Value, out.Asset
					}
					unblindedData, err := confidential.UnblindOutputWithKey(out, addrInfo.BlindingKey)
					if err != nil {
						s.warn(err, "failed to unblind output with given blind key")
						continue
					}
					var confirmedStatus domain.UtxoStatus
					if event.tx.Height > 0 {
						confirmedStatus = domain.UtxoStatus{
							BlockHash:   blockHash.String(),
							BlockTime:   blockTimestamp,
							BlockHeight: uint64(event.tx.Height),
						}
					}

					newUtxos = append(newUtxos, &domain.Utxo{
						UtxoKey: domain.UtxoKey{
							TxID: event.tx.Txid,
							VOut: uint32(i),
						},
						Asset:           elementsutil.TxIDFromBytes(unblindedData.Asset),
						Value:           unblindedData.Value,
						AssetCommitment: assetCommit,
						ValueCommitment: valueCommit,
						AssetBlinder:    unblindedData.AssetBlindingFactor,
						ValueBlinder:    unblindedData.ValueBlindingFactor,
						Nonce:           nonce,
						Script:          out.Script,
						AccountName:     event.account,
						ConfirmedStatus: confirmedStatus,
						RedeemScript:    addrInfo.RedeemScript,
					})
				}
			}
		}
	}

	chTx := s.getTxChannelByAccount(event.account)
	txHex, _ := tx.ToHex()

	var hash string
	var blockHeight uint64
	if blockHash != nil {
		hash = blockHash.String()
		blockHeight = uint64(event.tx.Height)
	}

	go func() {
		chTx <- &domain.Transaction{
			TxID:        event.tx.Txid,
			TxHex:       txHex,
			BlockHash:   hash,
			BlockHeight: blockHeight,
			Accounts:    map[string]struct{}{event.account: {}},
		}
	}()

	chUtxos := s.getUtxoChannelByAccount(event.account)
	if len(spentUtxos) > 0 {
		go func() { chUtxos <- spentUtxos }()
	}

	if len(confirmedUtxos) > 0 {
		go func() { chUtxos <- confirmedUtxos }()
	}

	if len(newUtxos) > 0 {
		go func() { chUtxos <- newUtxos }()
	}
}

func (s *service) getAddressByScriptHash(account, scriptHash string) *domain.AddressInfo {
	s.lock.RLock()
	defer s.lock.RUnlock()

	info, ok := s.accountAddressesByScriptHash[account][scriptHash]
	if !ok {
		return nil
	}
	return &info
}

func (s *service) setAddressesByScriptHash(account string, addresses []domain.AddressInfo) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if _, ok := s.accountAddressesByScriptHash[account]; !ok {
		s.accountAddressesByScriptHash[account] = make(map[string]domain.AddressInfo)
	}

	for _, addr := range addresses {
		s.accountAddressesByScriptHash[account][calcScriptHash(addr.Script)] = addr
	}
}
