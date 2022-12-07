package electrum_scanner

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"github.com/vulpemventures/go-elements/elementsutil"
	"github.com/vulpemventures/go-elements/transaction"
	"github.com/vulpemventures/ocean/internal/core/domain"
)

type wsClient struct {
	conn      *websocket.Conn
	nextId    uint64
	chHandler *chHandler
	chainTip  blockInfo
	lock      *sync.RWMutex

	log  func(format string, a ...interface{})
	warn func(err error, format string, a ...interface{})
}

func newWSClient(addr string) (electrumClient, error) {
	conn, _, err := websocket.DefaultDialer.Dial(addr, nil)
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

	return &wsClient{
		conn:      conn,
		nextId:    0,
		chHandler: newChHandler(),
		lock:      &sync.RWMutex{},
		log:       logFn,
		warn:      warnFn,
	}, nil
}

func (c *wsClient) listen() {
	var incompleteResp []byte
	for {
		var resp response
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseAbnormalClosure, websocket.CloseNoStatusReceived) {
				log.WithError(err).Fatal("connection dropped")
			}
			if errors.Is(err, net.ErrClosed) {
				return
			}
		}
		for _, m := range bytes.Split(msg, []byte{delim}) {
			if len(m) == 0 {
				continue
			}

			if len(incompleteResp) > 0 {
				m = append(incompleteResp, m...)
			}

			if err := json.Unmarshal(m, &resp); err != nil {
				incompleteResp = m
				continue
			}

			incompleteResp = make([]byte, 0)
			if err := resp.error(); err != nil {
				c.warn(err, "got response error from socket")
				continue
			}

			if len(resp.Method) > 0 {
				switch resp.Method {
				case "blockchain.scripthash.subscribe":
					scriptHash := resp.Params.([]interface{})[0].(string)
					account := c.chHandler.getAccountByScriptHash(scriptHash)
					chReports := c.chHandler.getChReportsForAccount(account)

					go func() { chReports <- accountReport{account, scriptHash} }()
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
}

func (c *wsClient) close() {
	c.conn.Close()
	c.chHandler.clear()
}

func (c *wsClient) subscribeForBlocks() {
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

func (c *wsClient) subscribeForAccount(
	accountName string, addresses []domain.AddressInfo,
) (chan accountReport, map[string][]txInfo) {
	history := make(map[string][]txInfo)
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

	if c.chHandler.getChReportsForAccount(accountName) == nil {
		c.chHandler.addChReportForAccount(accountName)
	}
	return c.chHandler.getChReportsForAccount(accountName), history
}

func (c *wsClient) unsubscribeForAccount(accountName string) {
	// hashes := c.chHandler.getAccountScriptHashes(accountName)
	// for _, scriptHash := range hashes {
	// 	c.unsubscribeForScript(accountName, scriptHash)
	// }
	c.chHandler.clearAccount(accountName)
}

func (c *wsClient) getScriptHashHistory(scriptHash string) ([]txInfo, error) {
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

func (c *wsClient) getLatestBlock() ([]byte, uint32, error) {
	return c.chainTip.hash()[:], uint32(c.chainTip.Height), nil
}

func (c *wsClient) getBlockInfo(height uint32) (*chainhash.Hash, int64, error) {
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

func (c *wsClient) getTx(txid string) (*transaction.Transaction, error) {
	resp, err := c.request("blockchain.transaction.get", txid)
	if err != nil {
		return nil, err
	}
	if err := resp.error(); err != nil {
		return nil, err
	}

	return transaction.NewTxFromHex(resp.Result.(string))
}

func (c *wsClient) getUtxos(outpoints []domain.Utxo) ([]domain.Utxo, error) {
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

func (c *wsClient) broadcastTx(txHex string) (string, error) {
	resp, err := c.request("blockchain.transaction.broadcast", txHex)
	if err != nil {
		return "", err
	}
	if err := resp.error(); err != nil {
		return "", err
	}
	return resp.Result.(string), nil
}

func (c *wsClient) subscribeForScript(accountName, scriptHash string) error {
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

// func (c *wsClient) unsubscribeForScript(accountName, scriptHash string) {
// 	req := c.newRequest("blockchain.scripthash.unsubscribe", scriptHash)
// 	buf, _ := json.Marshal(req)
// 	if err := c.conn.WriteMessage(websocket.TextMessage, buf); err != nil {
// 		c.warn(
// 			err, "failed to subscribe for script %s of account %s",
// 			scriptHash, accountName,
// 		)
// 		return
// 	}
// }

func (c *wsClient) request(method string, params ...interface{}) (*response, error) {
	req := c.newRequest(method, params...)
	if err := c.conn.WriteJSON(req); err != nil {
		c.warn(err, "failed to send request for method %s", method)
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

func (c *wsClient) newRequest(method string, params ...interface{}) request {
	params = append([]interface{}{}, params...)
	return request{atomic.AddUint64(&c.nextId, 1), method, params}
}

func (c *wsClient) updateChainTip(tip blockInfo) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.chainTip = tip
}
