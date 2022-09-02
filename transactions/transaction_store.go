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
	byPktID   map[uint16]Transaction
	byPktType map[pkts1.MessageType]Transaction
}

// NewTransactionStore creates a new transaction store.
func NewTransactionStore() *TransactionStore {
	return &TransactionStore{
		byPktID:   make(map[uint16]Transaction),
		byPktType: make(map[pkts1.MessageType]Transaction),
	}
}

// Store inserts a new transaction to the store by the MessageID.
func (ts *TransactionStore) Store(pktID uint16, transaction Transaction) {
	ts.Lock()
	defer ts.Unlock()
	ts.byPktID[pktID] = transaction
}

// StoreByType inserts a new transaction to the store by the MessageType.
func (ts *TransactionStore) StoreByType(pktType pkts1.MessageType, transaction Transaction) {
	ts.Lock()
	defer ts.Unlock()
	ts.byPktType[pktType] = transaction
}

// Get retrieves a transaction from the store by the MessageID.
func (ts *TransactionStore) Get(pktID uint16) (Transaction, bool) {
	ts.RLock()
	defer ts.RUnlock()
	transaction, ok := ts.byPktID[pktID]
	return transaction, ok
}

// GetByType retrieves a transaction from the store by the MessageType.
func (ts *TransactionStore) GetByType(pktType pkts1.MessageType) (Transaction, bool) {
	ts.Lock()
	defer ts.Unlock()
	transaction, ok := ts.byPktType[pktType]
	return transaction, ok
}

// Delete removes a transaction from the store by the MessageID.
func (ts *TransactionStore) Delete(pktID uint16) {
	ts.Lock()
	defer ts.Unlock()
	delete(ts.byPktID, pktID)
}

// DeleteByType removes a transaction from the store by the MessageType.
func (ts *TransactionStore) DeleteByType(pktType pkts1.MessageType) {
	ts.Lock()
	defer ts.Unlock()
	delete(ts.byPktType, pktType)
}
