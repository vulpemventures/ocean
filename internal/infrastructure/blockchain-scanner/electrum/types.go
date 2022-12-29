package electrum_scanner

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/vulpemventures/go-elements/block"
)

type txInfo struct {
	Txid   string `json:"tx_hash"`
	Height int64  `json:"height"`
}

type blockInfo struct {
	Header string `json:"hex"`
	Height uint64 `json:"height"`
}

func (i blockInfo) hash() *chainhash.Hash {
	header := i.header()
	if header == nil {
		return nil
	}

	hash, _ := header.Hash()
	return &hash
}

func (i blockInfo) timestamp() int64 {
	header := i.header()
	if header == nil {
		return -1
	}

	return int64(header.Timestamp)
}

func (i blockInfo) header() *block.Header {
	if i.Header == "" {
		return nil
	}

	buf, _ := hex.DecodeString(i.Header)
	header, _ := block.DeserializeHeader(bytes.NewBuffer(buf))
	return header
}

type request struct {
	Id     uint64        `json:"id"`
	Method string        `json:"method"`
	Params []interface{} `json:"params"`
}

type response struct {
	Id     uint64      `json:"id,omitempty"`
	Result interface{} `json:"result,omitempty"`
	Method string      `json:"method,omitempty"`
	Params interface{} `json:"params,omitempty"`
	Error  interface{} `json:"error,omitempty"`
}

func (r response) error() error {
	if r.Error == nil {
		return nil
	}

	if err, ok := r.Error.(string); ok {
		return fmt.Errorf(err)
	}

	buf, _ := json.Marshal(r.Error)
	var err responseErr
	json.Unmarshal(buf, &err)
	return err.Error()
}

type responseErr struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e responseErr) Error() error {
	if len(e.Message) <= 0 {
		return nil
	}
	return fmt.Errorf("code: %d, message: %s", e.Code, e.Message)
}

type accountReport struct {
	account    string
	scriptHash string
}

type chHandler struct {
	lock               *sync.RWMutex
	accountByScript    map[string]string
	scriptsByAccount   map[string][]string
	chReportsByAccount map[string]chan accountReport
	chReportsByReqId   map[uint32]chan response
}

func newChHandler() *chHandler {
	return &chHandler{
		lock:               &sync.RWMutex{},
		chReportsByAccount: make(map[string]chan accountReport),
		chReportsByReqId:   make(map[uint32]chan response),
		accountByScript:    make(map[string]string),
		scriptsByAccount:   make(map[string][]string),
	}
}

func (h *chHandler) addChReportForAccount(account string) {
	h.lock.Lock()
	defer h.lock.Unlock()

	if _, ok := h.chReportsByAccount[account]; ok {
		return
	}

	h.chReportsByAccount[account] = make(chan accountReport)
}

func (h *chHandler) addAccountScriptHash(account, scriptHash string) {
	h.lock.Lock()
	defer h.lock.Unlock()

	if _, ok := h.accountByScript[scriptHash]; ok {
		return
	}

	h.accountByScript[scriptHash] = account
	h.scriptsByAccount[account] = append(h.scriptsByAccount[account], scriptHash)
}

func (h *chHandler) addRequests(reqs []request) {
	h.lock.Lock()
	defer h.lock.Unlock()

	for _, req := range reqs {
		id := uint32(req.Id)
		if _, ok := h.chReportsByReqId[id]; ok {
			continue
		}

		h.chReportsByReqId[id] = make(chan response)
	}
}

func (h *chHandler) getAccountByScriptHash(script string) string {
	h.lock.RLock()
	defer h.lock.RUnlock()

	return h.accountByScript[script]
}

func (h *chHandler) getChReportsForAccount(account string) chan accountReport {
	h.lock.RLock()
	defer h.lock.RUnlock()

	return h.chReportsByAccount[account]
}

func (h *chHandler) getChReportsForReqId(id uint32) chan response {
	h.lock.RLock()
	defer h.lock.RUnlock()

	return h.chReportsByReqId[id]
}

func (h *chHandler) clearAccount(account string) {
	h.lock.Lock()
	defer h.lock.Unlock()

	if chReports, ok := h.chReportsByAccount[account]; ok {
		close(chReports)
	}
	delete(h.chReportsByAccount, account)
	delete(h.scriptsByAccount, account)
}

func (h *chHandler) clearRequest(id uint32) {
	h.lock.Lock()
	defer h.lock.Unlock()

	if chReports, ok := h.chReportsByReqId[id]; ok {
		close(chReports)
	}
	delete(h.chReportsByReqId, id)
}

func (h *chHandler) clear() {
	h.lock.RLock()
	defer h.lock.RUnlock()

	for _, ch := range h.chReportsByAccount {
		close(ch)
	}
	for _, ch := range h.chReportsByReqId {
		close(ch)
	}
}

type reportHandler struct {
	locker      *sync.Mutex
	isLocked    bool
	chReports   chan accountReport
	reportQueue []accountReport
}

func (h *reportHandler) lock() {
	h.locker.Lock()
	defer h.locker.Unlock()

	h.isLocked = true
}

func (h *reportHandler) unlock() {
	h.locker.Lock()
	defer h.locker.Unlock()

	for i := range h.reportQueue {
		report := h.reportQueue[i]
		go func(report accountReport) { h.chReports <- report }(report)
	}

	h.reportQueue = make([]accountReport, 0)
	h.isLocked = false
}

func (h *reportHandler) sendReport(report accountReport) {
	if h.isLocked {
		h.reportQueue = append(h.reportQueue, report)
		return
	}

	go func() { h.chReports <- report }()
}
