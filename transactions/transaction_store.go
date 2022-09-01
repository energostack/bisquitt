package transactions

import (
	"sync"

	pkts "github.com/energomonitor/bisquitt/packets"
)

// TransactionStore stores transactions by MessageID or PacketType
// (for packets types without MessageID). The typical usage is to store
// transactions in progress.
//
// It's safe for concurrent use.
type TransactionStore struct {
	sync.RWMutex
	bypktID   map[uint16]Transaction
	bypktType map[pkts.PacketType]Transaction
}

// NewTransactionStore creates a new transaction store.
func NewTransactionStore() *TransactionStore {
	return &TransactionStore{
		bypktID:   make(map[uint16]Transaction),
		bypktType: make(map[pkts.PacketType]Transaction),
	}
}

// Store inserts a new transaction to the store by the MessageID.
func (ts *TransactionStore) Store(pktID uint16, transaction Transaction) {
	ts.Lock()
	defer ts.Unlock()
	ts.bypktID[pktID] = transaction
}

// StoreByType inserts a new transaction to the store by the PacketType.
func (ts *TransactionStore) StoreByType(pktType pkts.PacketType, transaction Transaction) {
	ts.Lock()
	defer ts.Unlock()
	ts.bypktType[pktType] = transaction
}

// Get retrieves a transaction from the store by the MessageID.
func (ts *TransactionStore) Get(pktID uint16) (Transaction, bool) {
	ts.RLock()
	defer ts.RUnlock()
	transaction, ok := ts.bypktID[pktID]
	return transaction, ok
}

// GetByType retrieves a transaction from the store by the PacketType.
func (ts *TransactionStore) GetByType(pktType pkts.PacketType) (Transaction, bool) {
	ts.Lock()
	defer ts.Unlock()
	transaction, ok := ts.bypktType[pktType]
	return transaction, ok
}

// Delete removes a transaction from the store by the MessageID.
func (ts *TransactionStore) Delete(pktID uint16) {
	ts.Lock()
	defer ts.Unlock()
	delete(ts.bypktID, pktID)
}

// DeleteByType removes a transaction from the store by the PacketType.
func (ts *TransactionStore) DeleteByType(pktType pkts.PacketType) {
	ts.Lock()
	defer ts.Unlock()
	delete(ts.bypktType, pktType)
}
