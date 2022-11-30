package electrum_scanner

import (
	"sync"
)

const (
	txUnconfirmed dbEventType = iota
	txConfirmed
)

type dbEventType int

func (t dbEventType) String() string {
	switch t {
	case txUnconfirmed:
		return "TX_UNCONFIRMED"
	case txConfirmed:
		return "TX_CONFIRMED"
	default:
		return "UNKNOWN"
	}
}

type dbEvent struct {
	eventType  dbEventType
	tx         txInfo
	account    string
	scriptHash string
}

type db struct {
	lock   *sync.RWMutex
	chLock *sync.RWMutex

	txHistoryByAccount map[string]map[string]int64
	eventHandler       func(dbEvent)
	chEvents           chan dbEvent
}

func newDb() *db {
	db := &db{
		lock:   &sync.RWMutex{},
		chLock: &sync.RWMutex{},

		txHistoryByAccount: make(map[string]map[string]int64),
		chEvents:           make(chan dbEvent),
	}

	go db.listen()

	return db
}

func (d *db) updateAccountTxHistory(account, scriptHash string, newHistory []txInfo) {
	d.lock.Lock()
	defer d.lock.Unlock()

	if _, ok := d.txHistoryByAccount[account]; !ok {
		d.txHistoryByAccount[account] = make(map[string]int64)
	}

	for _, tx := range newHistory {
		prevHistory := d.txHistoryByAccount
		// unconfirmed txs have height 0 or -1, while those confirmed have height
		// equals to the one of the block in which they are contained.
		// If the tx is stored in the db and is confirmed, we don't have nothing
		// to do and we can skip to the next tx of the given history.
		if height := prevHistory[account][tx.Txid]; height > 0 {
			continue
		}

		d.txHistoryByAccount[account][tx.Txid] = tx.Height

		eventType := txUnconfirmed
		if tx.Height > 0 {
			eventType = txConfirmed
		}
		event := dbEvent{eventType, tx, account, scriptHash}
		go d.publishEvent(event)
	}
}

func (d *db) listen() {
	for event := range d.chEvents {
		if d.eventHandler != nil {
			d.eventHandler(event)
		}
	}
}

func (d *db) publishEvent(event dbEvent) {
	d.chLock.Lock()
	defer d.chLock.Unlock()

	go func() { d.chEvents <- event }()
}

func (d *db) registerEventHandler(handler func(event dbEvent)) {
	d.chLock.Lock()
	defer d.chLock.Unlock()

	if d.eventHandler == nil {
		d.eventHandler = handler
	}
}

func (d *db) close() {
	close(d.chEvents)
}
