package transactions

import (
	"sync"

	pkts1 "github.com/energomonitor/bisquitt/packets1"
)

// TransactionStore stores transactions by MessageID or MessageType
// (for packets types without MessageID). The typical usage is to store
// transactions in progress.
//
// It's safe for concurrent use.
type TransactionStore struct {
	sync.RWMutex
	byMsgID   map[uint16]Transaction
	byMsgType map[pkts1.MessageType]Transaction
}

// NewTransactionStore creates a new transaction store.
func NewTransactionStore() *TransactionStore {
	return &TransactionStore{
		byMsgID:   make(map[uint16]Transaction),
		byMsgType: make(map[pkts1.MessageType]Transaction),
	}
}

// Store inserts a new transaction to the store by the MessageID.
func (ts *TransactionStore) Store(msgID uint16, transaction Transaction) {
	ts.Lock()
	defer ts.Unlock()
	ts.byMsgID[msgID] = transaction
}

// StoreByType inserts a new transaction to the store by the MessageType.
func (ts *TransactionStore) StoreByType(msgType pkts1.MessageType, transaction Transaction) {
	ts.Lock()
	defer ts.Unlock()
	ts.byMsgType[msgType] = transaction
}

// Get retrieves a transaction from the store by the MessageID.
func (ts *TransactionStore) Get(msgID uint16) (Transaction, bool) {
	ts.RLock()
	defer ts.RUnlock()
	transaction, ok := ts.byMsgID[msgID]
	return transaction, ok
}

// GetByType retrieves a transaction from the store by the MessageType.
func (ts *TransactionStore) GetByType(msgType pkts1.MessageType) (Transaction, bool) {
	ts.Lock()
	defer ts.Unlock()
	transaction, ok := ts.byMsgType[msgType]
	return transaction, ok
}

// Delete removes a transaction from the store by the MessageID.
func (ts *TransactionStore) Delete(msgID uint16) {
	ts.Lock()
	defer ts.Unlock()
	delete(ts.byMsgID, msgID)
}

// DeleteByType removes a transaction from the store by the MessageType.
func (ts *TransactionStore) DeleteByType(msgType pkts1.MessageType) {
	ts.Lock()
	defer ts.Unlock()
	delete(ts.byMsgType, msgType)
}
