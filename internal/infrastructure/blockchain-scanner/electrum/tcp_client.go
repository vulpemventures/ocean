package electrum_scanner

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/vulpemventures/go-elements/elementsutil"
	"github.com/vulpemventures/go-elements/transaction"
	"github.com/vulpemventures/ocean/internal/core/domain"
)

const delim = byte('\n')

type tcpClient struct {
	addr           string
	conn           net.Conn
	nextId         uint64
	chHandler      *chHandler
	chainTip       blockInfo
	reportHandlers map[string]*reportHandler
	chQuit         chan struct{}
	subscriptions  []request

	tipLock  *sync.RWMutex
	sendLock *sync.RWMutex

	log  func(format string, a ...interface{})
	warn func(err error, format string, a ...interface{})
}

func newTCPClient(addr string) (electrumClient, error) {
	conn, err := getConn(addr)
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

	svc := &tcpClient{
		addr:           addr,
		conn:           conn,
		nextId:         0,
		chHandler:      newChHandler(),
		reportHandlers: make(map[string]*reportHandler, 0),
		chQuit:         make(chan struct{}),
		subscriptions:  make([]request, 0),
		tipLock:        &sync.RWMutex{},
		sendLock:       &sync.RWMutex{},
		log:            logFn,
		warn:           warnFn,
	}

	go svc.keepAliveConnection()

	return svc, nil
}

func (c *tcpClient) listen() {
	conn := bufio.NewReader(c.conn)
	for {
		var resp response
		bytes, err := conn.ReadBytes(delim)
		if err != nil {
			if errors.Is(err, net.ErrClosed) || errors.Is(err, io.EOF) ||
				errors.Is(err, syscall.ECONNRESET) {
				c.log("connection with server dropped, attempting to reconnect...")
				c.reconnect()
				return
			}
			log.WithError(err).Fatal("failed to read message from socket")
		}

		if err := json.Unmarshal(bytes, &resp); err != nil {
			c.warn(err, "failed to parse received message")
			continue
		}

		if len(resp.Method) > 0 {
			switch resp.Method {
			case "blockchain.scripthash.subscribe":
				scriptHash := resp.Params.([]interface{})[0].(string)
				account := c.chHandler.getAccountByScriptHash(scriptHash)
				report := accountReport{account, scriptHash}
				c.reportHandlers[account].sendReport(report)
				continue
			case "blockchain.headers.subscribe":
				buf, _ := json.Marshal(resp.Params.([]interface{})[0])
				var block blockInfo
				json.Unmarshal(buf, &block)

				c.updateChainTip(block)
				continue
			}
		}

		chReports := c.chHandler.getChReportsForReqId(uint32(resp.Id))
		go func() { chReports <- resp }()
	}
}

func (c *tcpClient) keepAliveConnection() {
	t := time.NewTicker(1 * time.Minute)

	for {
		select {
		case <-t.C:
			if _, err := c.request("server.ping"); err != nil {
				log.WithError(err).Error("scanner: failed to keep connection alive")
			}
		case <-c.chQuit:
			return
		}
	}
}

func (c *tcpClient) reconnect() {
	// stop sending ping messages.
	c.chQuit <- struct{}{}

	// re-establish the connection with electrum server.
	conn, err := getConn(c.addr)
	if err != nil {
		log.WithError(err).Fatal("failed to reconnect to server")
	}

	c.conn = conn

	// restart listening over tcp socket and sending ping messages.
	go c.listen()
	go c.keepAliveConnection()

	// restore all subscriptions
	c.log("restoring %d subscriptions...", len(c.subscriptions))
	responses, err := c.batchRequests(c.subscriptions)
	if err != nil {
		log.WithError(err).Fatal("failed to restore subscriptions after reconnection")
	}
	for _, resp := range responses {
		if err := resp.error(); err != nil {
			log.WithError(err).Fatal("failed to restore subscriptions after reconnection")
		}
	}

	c.log("connection with server restored")
}

func (c *tcpClient) close() {
	c.conn.Close()
	c.chHandler.clear()
	c.chQuit <- struct{}{}
	close(c.chQuit)
}

func (c *tcpClient) subscribeForBlocks() {
	resp, err := c.request("blockchain.headers.subscribe")
	if err != nil {
		c.warn(err, "failed to subscribe for new blocks")
		return
	}

	buf, _ := json.Marshal(resp.Result)
	var block blockInfo
	json.Unmarshal(buf, &block)

	c.updateChainTip(block)

	c.subscriptions = append(
		c.subscriptions, c.newJSONRequest("blockchain.headers.subscribe"),
	)
}

func (c *tcpClient) subscribeForAccount(
	accountName string, addresses []domain.AddressInfo,
) (chan accountReport, map[string][]txInfo) {
	if c.chHandler.getChReportsForAccount(accountName) == nil {
		c.chHandler.addChReportForAccount(accountName)
	}

	chReports := c.chHandler.getChReportsForAccount(accountName)
	if _, ok := c.reportHandlers[accountName]; !ok {
		c.reportHandlers[accountName] = &reportHandler{
			locker:      &sync.Mutex{},
			chReports:   chReports,
			reportQueue: make([]accountReport, 0),
		}
	}

	scriptHashes := make([]string, 0, len(addresses))
	c.reportHandlers[accountName].lock()
	defer c.reportHandlers[accountName].unlock()

	for _, info := range addresses {
		scriptHashes = append(scriptHashes, calcScriptHash(info.Script))
		c.log(
			"start watching address %s for account %s",
			info.DerivationPath, accountName,
		)
	}

	if err := c.subscribeForScripts(accountName, scriptHashes); err != nil {
		c.warn(
			err, "failed to subscribe for scripts of account %s", accountName,
		)
	}

	history, err := c.getScriptHashesHistory(scriptHashes)
	if err != nil {
		c.warn(
			err, "failed to get get tx history for watched addresses of account %s",
			accountName,
		)
	}

	return chReports, history
}

func (c *tcpClient) unsubscribeForAccount(accountName string) {
	// TODO: uncomment this if ElectrumX servers will support this in the future.
	// hashes := c.chHandler.getAccountScriptHashes(accountName)
	// for _, scriptHash := range hashes {
	// 	c.unsubscribeForScript(accountName, scriptHash)
	// }
	c.chHandler.clearAccount(accountName)
}

func (c *tcpClient) subscribeForScripts(
	accountName string, scriptHashes []string,
) error {
	reqs := make([]request, 0, len(scriptHashes))
	for _, scriptHash := range scriptHashes {
		reqs = append(reqs, c.newJSONRequest(
			"blockchain.scripthash.subscribe", scriptHash),
		)
	}

	responses, err := c.batchRequests(reqs)
	if err != nil {
		return err
	}
	for _, resp := range responses {
		if err := resp.error(); err != nil {
			return err
		}
	}

	for _, scriptHash := range scriptHashes {
		c.chHandler.addAccountScriptHash(accountName, scriptHash)
	}
	c.subscriptions = append(c.subscriptions, reqs...)
	return nil
}

func (c *tcpClient) getScriptHashesHistory(scriptHashes []string) (map[string][]txInfo, error) {
	reqs := make([]request, 0, len(scriptHashes))
	scriptHashById := make(map[uint64]string)
	for _, scriptHash := range scriptHashes {
		req := c.newJSONRequest("blockchain.scripthash.get_history", scriptHash)
		reqs = append(reqs, req)
		scriptHashById[req.Id] = scriptHash
	}
	responses, err := c.batchRequests(reqs)
	if err != nil {
		return nil, err
	}

	allHistory := make(map[string][]txInfo)
	for _, resp := range responses {
		if err := resp.error(); err != nil {
			continue
		}

		scriptHash := scriptHashById[resp.Id]
		history := make([]txInfo, 0)
		buf, _ := json.Marshal(resp.Result)
		json.Unmarshal(buf, &history)

		allHistory[scriptHash] = append(allHistory[scriptHash], history...)
	}

	return allHistory, nil
}

func (c *tcpClient) getLatestBlock() ([]byte, uint32, error) {
	return c.chainTip.hash()[:], uint32(c.chainTip.Height), nil
}

func (c *tcpClient) getBlocksInfo(heights []uint32) ([]blockInfo, error) {
	reqs := make([]request, 0, len(heights))
	heightByReqId := make(map[uint64]uint32)
	for _, height := range heights {
		req := c.newJSONRequest("blockchain.block.header", height)
		reqs = append(reqs, req)
		heightByReqId[req.Id] = height
	}

	responses, err := c.batchRequests(reqs)
	if err != nil {
		return nil, err
	}

	blocks := make([]blockInfo, 0, len(heights))
	for _, resp := range responses {
		if err := resp.error(); err != nil {
			return nil, err
		}
		header := resp.Result.(string)
		height := heightByReqId[resp.Id]
		blocks = append(blocks, blockInfo{Header: header, Height: uint64(height)})
	}

	return blocks, nil
}

func (c *tcpClient) getTxs(txids []string) ([]*transaction.Transaction, error) {
	reqs := make([]request, 0, len(txids))
	for _, txid := range txids {
		reqs = append(reqs, c.newJSONRequest("blockchain.transaction.get", txid))
	}
	responses, err := c.batchRequests(reqs)
	if err != nil {
		return nil, err
	}

	txs := make([]*transaction.Transaction, 0, len(txids))
	for _, resp := range responses {
		if err := resp.error(); err != nil {
			return nil, err
		}
		tx, err := transaction.NewTxFromHex(resp.Result.(string))
		if err != nil {
			return nil, err
		}
		txs = append(txs, tx)
	}

	return txs, nil
}

func (c *tcpClient) getUtxos(outpoints []domain.Utxo) ([]domain.Utxo, error) {
	utxos := make([]domain.Utxo, 0, len(outpoints))
	for _, u := range outpoints {
		txs, err := c.getTxs([]string{u.TxID})
		if err != nil {
			return nil, err
		}
		tx := txs[0]

		prevout := tx.Outputs[u.VOut]
		var value uint64
		var asset string
		var valueCommit, assetCommit, nonce []byte
		if prevout.IsConfidential() {
			valueCommit, assetCommit, nonce = prevout.Value, prevout.Asset, prevout.Nonce
		} else {
			value, _ = elementsutil.ValueFromBytes(prevout.Value)
			asset = elementsutil.AssetHashFromBytes(prevout.Asset)
		}
		utxos = append(utxos, domain.Utxo{
			UtxoKey:         u.Key(),
			Value:           value,
			Asset:           asset,
			ValueCommitment: valueCommit,
			AssetCommitment: assetCommit,
			Nonce:           nonce,
			Script:          prevout.Script,
			RangeProof:      prevout.RangeProof,
			SurjectionProof: prevout.SurjectionProof,
		})
	}
	return utxos, nil
}

func (c *tcpClient) broadcastTx(txHex string) (string, error) {
	resp, err := c.request("blockchain.transaction.broadcast", txHex)
	if err != nil {
		return "", err
	}
	if err := resp.error(); err != nil {
		return "", err
	}
	return resp.Result.(string), nil
}

func (c *tcpClient) request(method string, params ...interface{}) (*response, error) {
	req := c.newJSONRequest(method, params...)
	reqBytes, _ := json.Marshal(req)
	reqBytes = append(reqBytes, delim)
	if _, err := c.conn.Write(reqBytes); err != nil {
		return nil, err
	}

	c.chHandler.addRequests([]request{req})
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer func() {
		cancel()
		c.chHandler.clearRequest(uint32(req.Id))
	}()

	select {
	case resp := <-c.chHandler.getChReportsForReqId(uint32(req.Id)):
		return &resp, nil
	case <-ctx.Done():
		return nil, fmt.Errorf("request timed out")
	}
}

func (c *tcpClient) batchRequests(reqs []request) ([]response, error) {
	reqBytes := make([]byte, 0)
	for _, req := range reqs {
		buf, _ := json.Marshal(req)
		reqBytes = append(reqBytes, append(buf, delim)...)
	}

	if _, err := c.conn.Write(reqBytes); err != nil {
		return nil, err
	}

	c.chHandler.addRequests(reqs)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer func() {
		cancel()
		for _, req := range reqs {
			c.chHandler.clearRequest(uint32(req.Id))
		}
	}()

	lock := sync.Mutex{}
	responses := make([]response, 0, len(reqs))
	wg := sync.WaitGroup{}
	wg.Add(len(reqs))

	for i := range reqs {
		req := reqs[i]
		ch := c.chHandler.getChReportsForReqId(uint32(req.Id))
		go func(chResp chan response) {
			select {
			case resp := <-chResp:
				lock.Lock()
				responses = append(responses, resp)
				lock.Unlock()
			case <-ctx.Done():
				c.warn(nil, "request timed out")
			}
			wg.Done()
		}(ch)
	}
	wg.Wait()

	return responses, nil
}

func (c *tcpClient) newJSONRequest(method string, params ...interface{}) request {
	params = append([]interface{}{}, params...)
	return request{atomic.AddUint64(&c.nextId, 1), method, params}
}

func (c *tcpClient) updateChainTip(tip blockInfo) {
	c.tipLock.Lock()
	defer c.tipLock.Unlock()

	c.chainTip = tip
}
