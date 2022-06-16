package transactions

import (
	"sync"

	pkts "github.com/energomonitor/bisquitt/packets1"
)

// TransactionStore stores transactions by message ID or by message type
// (for message types without a message ID). The typical usage is to store
// transactions in progress.
//
// It's safe for concurrent use.
type TransactionStore struct {
	sync.RWMutex
	byMsgID   map[uint16]Transaction
	byMsgType map[pkts.MessageType]Transaction
}

// NewTransactionStore creates a new transaction store.
func NewTransactionStore() *TransactionStore {
	return &TransactionStore{
		byMsgID:   make(map[uint16]Transaction),
		byMsgType: make(map[pkts.MessageType]Transaction),
	}
}

// Store inserts a new transaction to the store by the message ID.
func (ts *TransactionStore) Store(msgID uint16, transaction Transaction) {
	ts.Lock()
	defer ts.Unlock()
	ts.byMsgID[msgID] = transaction
}

// StoreByType inserts a new transaction to the store by the message type.
func (ts *TransactionStore) StoreByType(msgType pkts.MessageType, transaction Transaction) {
	ts.Lock()
	defer ts.Unlock()
	ts.byMsgType[msgType] = transaction
}

// Get retrieves a transaction from the store by the message ID.
func (ts *TransactionStore) Get(msgID uint16) (Transaction, bool) {
	ts.RLock()
	defer ts.RUnlock()
	transaction, ok := ts.byMsgID[msgID]
	return transaction, ok
}

// GetByType retrieves a transaction from the store by the message type.
func (ts *TransactionStore) GetByType(msgType pkts.MessageType) (Transaction, bool) {
	ts.Lock()
	defer ts.Unlock()
	transaction, ok := ts.byMsgType[msgType]
	return transaction, ok
}

// Delete removes a transaction from the store by the message ID.
func (ts *TransactionStore) Delete(msgID uint16) {
	ts.Lock()
	defer ts.Unlock()
	delete(ts.byMsgID, msgID)
}

// DeleteByType removes a transaction from the store by the message type.
func (ts *TransactionStore) DeleteByType(msgType pkts.MessageType) {
	ts.Lock()
	defer ts.Unlock()
	delete(ts.byMsgType, msgType)
}
