package electrum_scanner

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	log "github.com/sirupsen/logrus"
	"github.com/vulpemventures/go-elements/elementsutil"
	"github.com/vulpemventures/go-elements/transaction"
	"github.com/vulpemventures/ocean/internal/core/domain"
)

const delim = byte('\n')

type tcpClient struct {
	conn         net.Conn
	nextId       uint64
	chHandler    *chHandler
	chainTip     blockInfo
	isSending    bool
	reportQueue  []accountReport
	chSendStatus chan bool
	chQuit       chan struct{}

	tipLock  *sync.RWMutex
	sendLock *sync.RWMutex

	log  func(format string, a ...interface{})
	warn func(err error, format string, a ...interface{})
}

func newTCPClient(addr string) (electrumClient, error) {
	split := strings.Split(addr, "://")
	proto, url := split[0], split[1]
	var conn net.Conn
	switch proto {
	case "tcp":
		c, err := net.Dial(proto, url)
		if err != nil {
			return nil, err
		}
		conn = c
	case "ssl":
		c, err := tls.Dial("tcp", url, nil)
		if err != nil {
			return nil, err
		}
		conn = c
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
		conn:         conn,
		nextId:       0,
		chHandler:    newChHandler(),
		reportQueue:  make([]accountReport, 0),
		chSendStatus: make(chan bool),
		chQuit:       make(chan struct{}),
		tipLock:      &sync.RWMutex{},
		sendLock:     &sync.RWMutex{},
		log:          logFn,
		warn:         warnFn,
	}

	go svc.listenSendStatus()
	go svc.keepAliveConnection()

	return svc, nil
}

func (c *tcpClient) listen() {
	conn := bufio.NewReader(c.conn)
	for {
		var resp response
		bytes, err := conn.ReadBytes(delim)
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
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
				if c.isSending {
					c.reportQueue = append(c.reportQueue, report)
				} else {
					chReports := c.chHandler.getChReportsForAccount(account)
					go func() { chReports <- report }()
				}
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

func (c *tcpClient) close() {
	c.conn.Close()
	c.chHandler.clear()
	close(c.chSendStatus)
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
}

func (c *tcpClient) subscribeForAccount(
	accountName string, addresses []domain.AddressInfo,
) (chan accountReport, map[string][]txInfo) {
	history := make(map[string][]txInfo)
	c.setSendingStatus(true)
	for _, info := range addresses {
		scriptHash := calcScriptHash(info.Script)
		if err := c.subscribeForScript(accountName, scriptHash); err != nil {
			c.warn(
				err, "failed to subscribe for script %s of account %s",
				info.Script, accountName,
			)
			continue
		}

		c.log(
			"start watching address %s for account %s",
			info.DerivationPath, accountName,
		)

		addrHistory, err := c.getScriptHashHistory(scriptHash)
		if err != nil {
			continue
		}
		history[scriptHash] = addrHistory
	}
	c.setSendingStatus(false)

	if c.chHandler.getChReportsForAccount(accountName) == nil {
		c.chHandler.addChReportForAccount(accountName)
	}

	return c.chHandler.getChReportsForAccount(accountName), history
}

func (c *tcpClient) unsubscribeForAccount(accountName string) {
	// TODO: uncomment this if ElectrumX servers will support this in the future.
	// hashes := c.chHandler.getAccountScriptHashes(accountName)
	// for _, scriptHash := range hashes {
	// 	c.unsubscribeForScript(accountName, scriptHash)
	// }
	c.chHandler.clearAccount(accountName)
}

func (c *tcpClient) subscribeForScript(accountName, scriptHash string) error {
	resp, err := c.request("blockchain.scripthash.subscribe", scriptHash)
	if err != nil {
		return err
	}
	if err := resp.error(); err != nil {
		return err
	}
	c.chHandler.addAccountScriptHash(accountName, scriptHash)
	return nil
}

// Not supported by common ElectrumX servers
// func (c *tcpClient) unsubscribeForScript(accountName, scriptHash string) {
// 	req := c.newJSONRequest("blockchain.scripthash.unsubscribe", scriptHash)
// 	reqBytes, _ := json.Marshal(req)
// 	if _, err := c.conn.Write(reqBytes); err != nil {
// 		c.warn(
// 			err, "failed to unsubscribe for script %s of account %s",
// 			scriptHash, accountName,
// 		)
// 		return
// 	}
// }

func (c *tcpClient) getScriptHashHistory(scriptHash string) ([]txInfo, error) {
	resp, err := c.request("blockchain.scripthash.get_history", scriptHash)
	if err != nil {
		return nil, err
	}
	if err := resp.error(); err != nil {
		return nil, err
	}

	buf, _ := json.Marshal(resp.Result)
	history := make([]txInfo, 0)
	json.Unmarshal(buf, &history)
	return history, nil
}

func (c *tcpClient) getLatestBlock() ([]byte, uint32, error) {
	return c.chainTip.hash()[:], uint32(c.chainTip.Height), nil
}

func (c *tcpClient) getBlockInfo(height uint32) (*chainhash.Hash, int64, error) {
	resp, err := c.request("blockchain.block.header", height)
	if err != nil {
		return nil, -1, err
	}
	if err := resp.error(); err != nil {
		return nil, -1, err
	}

	block := blockInfo{Header: resp.Result.(string)}
	return block.hash(), block.timestamp(), nil
}

func (c *tcpClient) getTx(txid string) (*transaction.Transaction, error) {
	resp, err := c.request("blockchain.transaction.get", txid)
	if err != nil {
		return nil, err
	}
	if err := resp.error(); err != nil {
		return nil, err
	}

	return transaction.NewTxFromHex(resp.Result.(string))
}

func (c *tcpClient) getUtxos(outpoints []domain.Utxo) ([]domain.Utxo, error) {
	utxos := make([]domain.Utxo, 0, len(outpoints))
	for _, u := range outpoints {
		tx, err := c.getTx(u.TxID)
		if err != nil {
			return nil, err
		}

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

	c.chHandler.addRequest(req)
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

func (c *tcpClient) newJSONRequest(method string, params ...interface{}) request {
	params = append([]interface{}{}, params...)
	return request{atomic.AddUint64(&c.nextId, 1), method, params}
}

func (c *tcpClient) updateChainTip(tip blockInfo) {
	c.tipLock.Lock()
	defer c.tipLock.Unlock()

	c.chainTip = tip
}

func (c *tcpClient) setSendingStatus(val bool) {
	c.sendLock.Lock()
	defer c.sendLock.Unlock()

	c.isSending = val
	c.chSendStatus <- val
}

func (c *tcpClient) listenSendStatus() {
	for isSending := range c.chSendStatus {
		if !isSending {
			for _, report := range c.reportQueue {
				chReports := c.chHandler.getChReportsForAccount(report.account)
				go func(chReport chan accountReport, report accountReport) {
					chReports <- report
				}(chReports, report)
			}
			c.reportQueue = make([]accountReport, 0)
		}
	}
}
