package transactions

import (
	"sync"

	msgs "github.com/energomonitor/bisquitt/messages"
)

// TransactionStore stores transactions by message ID or by message type
// (for message types without a message ID). The typical usage is to store
// transactions in progress.
//
// It's safe for concurrent use.
type TransactionStore struct {
	sync.RWMutex
	byMsgID   map[uint16]Transaction
	byMsgType map[msgs.MessageType]Transaction
}

// NewTransactionStore creates a new transaction store.
func NewTransactionStore() *TransactionStore {
	return &TransactionStore{
		byMsgID:   make(map[uint16]Transaction),
		byMsgType: make(map[msgs.MessageType]Transaction),
	}
}

// Store inserts a new transaction to the store by the message ID.
func (ts *TransactionStore) Store(msgID uint16, transaction Transaction) {
	ts.Lock()
	defer ts.Unlock()
	ts.byMsgID[msgID] = transaction
}

// StoreByType inserts a new transaction to the store by the message type.
func (ts *TransactionStore) StoreByType(msgType msgs.MessageType, transaction Transaction) {
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
func (ts *TransactionStore) GetByType(msgType msgs.MessageType) (Transaction, bool) {
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
func (ts *TransactionStore) DeleteByType(msgType msgs.MessageType) {
	ts.Lock()
	defer ts.Unlock()
	delete(ts.byMsgType, msgType)
}
