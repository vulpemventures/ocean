package electrum_scanner

import (
	"encoding/hex"
	"fmt"
	"strings"
	"sync"

	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	log "github.com/sirupsen/logrus"
	"github.com/vulpemventures/go-elements/confidential"
	"github.com/vulpemventures/go-elements/elementsutil"
	"github.com/vulpemventures/go-elements/network"
	"github.com/vulpemventures/go-elements/payment"
	"github.com/vulpemventures/go-elements/slip77"
	"github.com/vulpemventures/go-elements/transaction"
	"github.com/vulpemventures/ocean/internal/core/domain"
	"github.com/vulpemventures/ocean/internal/core/ports"
)

type service struct {
	client electrumClient
	db     *db

	lock *sync.RWMutex

	net                          *network.Network
	accountAddressesByScriptHash map[string]map[string]domain.AddressInfo
	utxoChannelByAccount         map[string]chan []*domain.Utxo
	txChannelByAccount           map[string]chan *domain.Transaction
	reportChannelByAccount       map[string]chan accountReport
	blocksByHeight               map[uint64]blockInfo

	log  func(format string, a ...interface{})
	warn func(err error, format string, a ...interface{})
}

type ServiceArgs struct {
	Addr    string
	Network *network.Network
}

func (a ServiceArgs) validate() error {
	if a.Addr == "" {
		return fmt.Errorf("missing ws endpoint")
	}
	if !a.withTCP() && !a.withWS() {
		return fmt.Errorf("invalid address: unknown protocol")
	}
	if a.Network == nil {
		return fmt.Errorf("missing network")
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
	blocksByHeight := make(map[uint64]blockInfo)

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
		client, db, lock, args.Network, accountAddressesByScriptHash,
		utxoChannelByAccount, txChannelByAccount, reportChannelByAccount,
		blocksByHeight, logFn, warnFn,
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
	accountName string, _ uint32, addresses []domain.AddressInfo,
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

func (s *service) RestoreAccount(
	accountIndex uint32, accountName, xpub string, masterBlindingKey []byte,
	_, addressesThreshold uint32,
) ([]domain.AddressInfo, []domain.AddressInfo, error) {
	masterKey, err := hdkeychain.NewKeyFromString(xpub)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid xpub: %s", err)
	}

	masterBlindKey, err := slip77.FromMasterKey(masterBlindingKey)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid master blinding key: %s", err)
	}

	externalAddresses := s.restoreAddressesForAccount(
		accountName, accountIndex, 0, masterKey, masterBlindKey, addressesThreshold,
	)
	internalAddresses := s.restoreAddressesForAccount(
		accountName, accountIndex, 1, masterKey, masterBlindKey, addressesThreshold,
	)

	return externalAddresses, internalAddresses, nil
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
	blocks, err := s.client.getBlocksInfo([]uint32{height})
	if err != nil {
		return nil, err
	}
	return blocks[0].hash()[:], nil
}

// GetUtxos is a sync function to get info about the utxos represented by
// given outpoints (UtxoKeys).
func (s *service) GetUtxos(utxos []domain.Utxo) ([]domain.Utxo, error) {
	return s.client.getUtxos(utxos)
}

func (s *service) GetUtxosForAddresses(
	addresses []domain.AddressInfo,
) ([]*domain.Utxo, error) {
	// Parse addresses into script hashes.
	scriptHashes := make([]string, 0, len(addresses))
	addressesByScriptHash := make(map[string]domain.AddressInfo)
	for _, addr := range addresses {
		scriptHash := calcScriptHash(addr.Script)
		scriptHashes = append(scriptHashes, scriptHash)
		addressesByScriptHash[scriptHash] = addr
	}

	// Retrieve tx history for all addresses.
	history, err := s.client.getScriptHashesHistory(scriptHashes)
	if err != nil {
		return nil, err
	}

	allTxsById := make(map[string]txInfo)
	for _, txHistory := range history {
		for _, txInfo := range txHistory {
			allTxsById[txInfo.Txid] = txInfo
		}
	}
	allTxs := make([]string, 0, len(allTxsById))
	for txid := range allTxsById {
		allTxs = append(allTxs, txid)
	}

	txs, err := s.client.getTxs(allTxs)
	if err != nil {
		return nil, err
	}

	// For every tx of the history, add all inputs to the list (map) of spent
	// utxos, and all outputs owned by any of the given addresses to the
	// list (map) of all utxos.
	spentUtxosByKey := make(map[domain.UtxoKey]struct{})
	utxosByScriptHash := make(map[string][]*domain.Utxo)
	for _, tx := range txs {
		for _, in := range tx.Inputs {
			spentUtxosByKey[domain.UtxoKey{
				TxID: elementsutil.TxIDFromBytes(in.Hash),
				VOut: in.Index,
			}] = struct{}{}
		}
		for i, out := range tx.Outputs {
			if len(out.Script) > 0 {
				scriptHash := calcScriptHash(hex.EncodeToString(out.Script))
				if _, ok := addressesByScriptHash[scriptHash]; ok {
					var value uint64
					var asset string
					var nonce, valueCommit, assetCommit []byte
					if out.IsConfidential() {
						nonce, valueCommit, assetCommit = out.Nonce, out.Value, out.Asset
					} else {
						value, _ = elementsutil.ValueFromBytes(out.Value)
						asset = elementsutil.AssetHashFromBytes(out.Asset)
					}
					utxosByScriptHash[scriptHash] = append(utxosByScriptHash[scriptHash], &domain.Utxo{
						UtxoKey: domain.UtxoKey{
							TxID: tx.TxHash().String(),
							VOut: uint32(i),
						},
						Value:           value,
						Asset:           asset,
						ValueCommitment: valueCommit,
						AssetCommitment: assetCommit,
						Script:          out.Script,
						Nonce:           nonce,
						RangeProof:      out.RangeProof,
						SurjectionProof: out.SurjectionProof,
					})
				}
			}
		}
	}

	// Loop over the list of all utxos to unblind'em if necessary.
	// Last thing to do is to reconstruct the spent and confirmed statuses of
	// the utxos. For this, we collect all blocks (height) for which we need to
	// fetch info from electrum.
	// We make use of 2 auxiliary maps to associate blocks to spent and confimed
	// utxos.
	spentUtxoBlocks := make(map[uint32][]domain.UtxoKey)
	utxoBlocks := make(map[uint32][]domain.UtxoKey)
	utxosByKey := make(map[domain.UtxoKey]*domain.Utxo)
	for scriptHash, utxos := range utxosByScriptHash {
		for _, utxo := range utxos {
			if _, ok := spentUtxosByKey[utxo.Key()]; ok {
				if txInfo, ok := allTxsById[utxo.TxID]; ok {
					spentUtxoBlocks[uint32(txInfo.Height)] = append(
						spentUtxoBlocks[uint32(txInfo.Height)], utxo.Key(),
					)
				} else {
					continue
				}
			}

			blockHeight := allTxsById[utxo.TxID].Height
			if blockHeight > 0 {
				utxoBlocks[uint32(blockHeight)] = append(
					utxoBlocks[uint32(blockHeight)], utxo.Key(),
				)
			}

			addr := addressesByScriptHash[scriptHash]
			asset := utxo.AssetCommitment
			if len(asset) == 0 {
				asset, _ = elementsutil.AssetHashToBytes(utxo.Asset)
			}
			value := utxo.ValueCommitment
			if len(value) == 0 {
				value, _ = elementsutil.ValueToBytes(utxo.Value)
			}
			unblindedData, _ := confidential.UnblindOutputWithKey(&transaction.TxOutput{
				Value:           value,
				Asset:           asset,
				Script:          utxo.Script,
				Nonce:           utxo.Nonce,
				RangeProof:      utxo.RangeProof,
				SurjectionProof: utxo.SurjectionProof,
			}, addr.BlindingKey)
			utxo.Value = unblindedData.Value
			utxo.Asset = elementsutil.TxIDFromBytes(unblindedData.Asset)
			utxo.ValueBlinder = unblindedData.ValueBlindingFactor
			utxo.AssetBlinder = unblindedData.AssetBlindingFactor
			utxo.RangeProof, utxo.SurjectionProof = nil, nil
			utxosByKey[utxo.Key()] = utxo
		}
	}

	// Merge the 2 auxiliary maps into a single one to prevent duplicated keys
	// and fetch info for all blocks.
	allUtxoBlocks := make(map[uint32]struct{})
	for height := range utxoBlocks {
		allUtxoBlocks[height] = struct{}{}
	}
	for height := range spentUtxoBlocks {
		allUtxoBlocks[height] = struct{}{}
	}
	blocks := make([]uint32, 0, len(allUtxoBlocks))
	for height := range allUtxoBlocks {
		blocks = append(blocks, height)
	}
	blocksInfo, err := s.client.getBlocksInfo(blocks)
	if err != nil {
		return nil, err
	}

	// Reconstruct utxo statuses with the fetched blocks info and update the
	// statuses of the spent and confirmed utxos.
	for _, block := range blocksInfo {
		status := domain.UtxoStatus{
			BlockHeight: block.Height,
			BlockHash:   block.hash().String(),
			BlockTime:   block.timestamp(),
		}
		for _, key := range utxoBlocks[uint32(block.Height)] {
			utxosByKey[key].ConfirmedStatus = status
		}
		for _, key := range spentUtxoBlocks[uint32(block.Height)] {
			utxosByKey[key].SpentStatus = status
		}
	}

	// Translate the auxiliary map into a list of utxos.
	utxos := make([]*domain.Utxo, 0, len(utxosByKey))
	for _, utxo := range utxosByKey {
		utxos = append(utxos, utxo)
	}
	return utxos, nil
}

// BroadcastTransaction sends the given raw tx (in hex string) over the
// network in order to be included in a later block of the Liquid blockchain.
func (s *service) BroadcastTransaction(txHex string) (string, error) {
	return s.client.broadcastTx(txHex)
}

// GetTransactions returns info about the given txids.
func (s *service) GetTransactions(txids []string) ([]domain.Transaction, error) {
	res, err := s.client.getTxs(txids)
	if err != nil {
		return nil, err
	}
	txs := make([]domain.Transaction, 0, len(res))
	for _, tx := range res {
		txid := tx.TxHash().String()
		scriptHash := calcScriptHash(hex.EncodeToString(tx.Outputs[0].Script))
		history, _ := s.client.getScriptHashesHistory([]string{scriptHash})
		var height int64
		for _, tx := range history[scriptHash] {
			if tx.Txid == txid {
				height = tx.Height
				break
			}
		}
		var blockhash string
		var blockheight uint64
		var blocktime int64
		if height > 0 {
			info, ok := s.blocksByHeight[uint64(height)]
			if ok {
				blockhash = info.hash().String()
				blockheight = info.Height
				blocktime = info.timestamp()
			} else {
				info, _ := s.client.getBlocksInfo([]uint32{uint32(height)})
				if len(info) > 0 {
					s.blocksByHeight[uint64(height)] = info[0]
					blockhash = info[0].hash().String()
					blockheight = info[0].Height
					blocktime = info[0].timestamp()
				}
			}
		}
		txHex, _ := tx.ToHex()
		txs = append(txs, domain.Transaction{
			TxID:        txid,
			TxHex:       txHex,
			BlockHash:   blockhash,
			BlockHeight: blockheight,
			BlockTime:   blocktime,
		})
	}
	return txs, nil
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
	txs, err := s.client.getTxs([]string{event.tx.Txid})
	if err != nil {
		s.warn(err, "failed to fetch tx for event %+v", event)
		return
	}
	tx := txs[0]

	newUtxos := make([]*domain.Utxo, 0)
	spentUtxos := make([]*domain.Utxo, 0)
	confirmedUtxos := make([]*domain.Utxo, 0)
	confirmedSpentUtxos := make([]*domain.Utxo, 0)
	var blockhash string
	var blocktime int64
	var blockheight uint64

	if event.tx.Height > 0 {
		blocks, err := s.client.getBlocksInfo(
			[]uint32{uint32(event.tx.Height)},
		)
		if err != nil {
			s.warn(err, "failed to fetch block %d", event.tx.Height)
			return
		}
		block := &blocks[0]
		blockhash = block.hash().String()
		blocktime = block.timestamp()
		blockheight = uint64(event.tx.Height)
	}

	if event.eventType == txConfirmed {
		for _, in := range tx.Inputs {
			// Let's try to fetch the input's prevout to check if the utxo belongs to
			// watched script and has to be added to the list of those spent with
			// confirmation.
			// If for any reason the prevout is not fetched, the utxo is added to the
			// list anyway, the receiver will ignore it.
			utxoKey := domain.UtxoKey{
				TxID: elementsutil.TxIDFromBytes(in.Hash),
				VOut: in.Index,
			}
			prevout := s.getPrevout(utxoKey)
			if prevout != nil {
				scriptHash := calcScriptHash(hex.EncodeToString(prevout.Script))
				addrInfo := s.getAddressByScriptHash(event.account, scriptHash)
				if addrInfo == nil {
					continue
				}
			}

			confirmedSpentUtxos = append(confirmedSpentUtxos, &domain.Utxo{
				UtxoKey: utxoKey,
				SpentStatus: domain.UtxoStatus{
					Txid:        event.tx.Txid,
					BlockHash:   blockhash,
					BlockTime:   blocktime,
					BlockHeight: blockheight,
				},
				AccountName: event.account,
			})
		}
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
							BlockHash:   blockhash,
							BlockTime:   blocktime,
							BlockHeight: blockheight,
						},
						AccountName: event.account,
					})
				}
			}
		}
	}

	if event.eventType == txAdded {
		for _, in := range tx.Inputs {
			// Let's try to fetch the input's prevout to check if the utxo belongs to
			// watched script and has to be added to the list of those spent.
			// If for any reason the prevout is not fetched, the utxo is added to the
			// list anyway, the receiver will ignore it.
			utxoKey := domain.UtxoKey{
				TxID: elementsutil.TxIDFromBytes(in.Hash),
				VOut: in.Index,
			}
			prevout := s.getPrevout(utxoKey)
			if prevout != nil {
				scriptHash := calcScriptHash(hex.EncodeToString(prevout.Script))
				addrInfo := s.getAddressByScriptHash(event.account, scriptHash)
				if addrInfo == nil {
					continue
				}
			}
			spentUtxos = append(spentUtxos, &domain.Utxo{
				UtxoKey: domain.UtxoKey{
					TxID: elementsutil.TxIDFromBytes(in.Hash),
					VOut: in.Index,
				},
				SpentStatus: domain.UtxoStatus{
					Txid:        event.tx.Txid,
					BlockHash:   blockhash,
					BlockTime:   blocktime,
					BlockHeight: blockheight,
				},
				AccountName: event.account,
			})
		}
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
							BlockHash:   blockhash,
							BlockTime:   blocktime,
							BlockHeight: blockheight,
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
					})
				}
			}
		}
	}

	chTx := s.getTxChannelByAccount(event.account)
	txHex, _ := tx.ToHex()

	go func() {
		chTx <- &domain.Transaction{
			TxID:        event.tx.Txid,
			TxHex:       txHex,
			BlockHash:   blockhash,
			BlockHeight: blockheight,
			BlockTime:   blocktime,
			Accounts:    map[string]struct{}{event.account: {}},
		}
	}()

	chUtxos := s.getUtxoChannelByAccount(event.account)
	if len(newUtxos) > 0 {
		go func() { chUtxos <- newUtxos }()
	}

	if len(spentUtxos) > 0 {
		go func() { chUtxos <- spentUtxos }()
	}

	if len(confirmedUtxos) > 0 {
		go func() { chUtxos <- confirmedUtxos }()
	}

	if len(confirmedSpentUtxos) > 0 {
		go func() { chUtxos <- confirmedSpentUtxos }()
	}
}

func (s *service) getAddressByScriptHash(
	account, scriptHash string,
) *domain.AddressInfo {
	s.lock.RLock()
	defer s.lock.RUnlock()

	info, ok := s.accountAddressesByScriptHash[account][scriptHash]
	if !ok {
		return nil
	}
	return &info
}

func (s *service) setAddressesByScriptHash(
	account string, addresses []domain.AddressInfo,
) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if _, ok := s.accountAddressesByScriptHash[account]; !ok {
		s.accountAddressesByScriptHash[account] = make(map[string]domain.AddressInfo)
	}

	for _, addr := range addresses {
		s.accountAddressesByScriptHash[account][calcScriptHash(addr.Script)] = addr
	}
}

func (s *service) restoreAddressesForAccount(
	accountName string, accountIndex, chain uint32,
	masterKey *hdkeychain.ExtendedKey, masterBlindKey *slip77.Slip77,
	addressesThaddressesThreshold uint32,
) []domain.AddressInfo {
	batchSize := int(addressesThaddressesThreshold)
	batchCounter := 0
	unusedAddressesCounter := 0
	hdNode, _ := masterKey.Derive(chain)
	restoredAddresses := make([]domain.AddressInfo, 0)

	for {
		if unusedAddressesCounter >= batchSize {
			break
		}

		scriptHashes := make([]string, 0, batchSize)
		addressesByScriptHash := make(map[string]domain.AddressInfo)

		for i := 0; i < batchSize; i++ {
			index := uint32(i + batchSize*batchCounter)
			key, _ := hdNode.Derive(index)
			pubkey, _ := key.ECPubKey()
			unconf := payment.FromPublicKey(pubkey, s.net, nil)
			blindingPrvkey, blindingPubkey, _ := masterBlindKey.DeriveKey(
				unconf.WitnessScript,
			)
			p2wpkh := payment.FromPublicKey(pubkey, s.net, blindingPubkey)
			addr, _ := p2wpkh.ConfidentialWitnessPubKeyHash()
			script := hex.EncodeToString(p2wpkh.WitnessScript)
			scriptHash := calcScriptHash(script)

			scriptHashes = append(scriptHashes, scriptHash)
			addressesByScriptHash[scriptHash] = domain.AddressInfo{
				Account:        accountName,
				Address:        addr,
				BlindingKey:    blindingPrvkey.Serialize(),
				DerivationPath: fmt.Sprintf("%d'/%d/%d", accountIndex, chain, index),
				Script:         script,
			}
		}

		history, _ := s.client.getScriptHashesHistory(scriptHashes)
		if len(history) <= 0 {
			break
		}

		for _, scriptHash := range scriptHashes {
			if txHistory := history[scriptHash]; len(txHistory) > 0 {
				unusedAddressesCounter = 0
				restoredAddresses = append(
					restoredAddresses, addressesByScriptHash[scriptHash],
				)
				continue
			}
			unusedAddressesCounter++
		}

		batchCounter++
	}

	return restoredAddresses
}

func (s *service) getPrevout(utxo domain.UtxoKey) *transaction.TxOutput {
	res, err := s.client.getTxs([]string{utxo.TxID})
	if err != nil {
		return nil
	}
	tx := res[0]
	if len(tx.Outputs) <= int(utxo.VOut) {
		return nil
	}
	return tx.Outputs[utxo.VOut]
}
