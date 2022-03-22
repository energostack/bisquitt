package transactions

import (
	"sync"
)

// Finally callback type.
type FinallyCallback func()

// TransactionBase is a basic Transaction interface implementation to be
// embedded in the specific transaction implementations. Besides the Transaction
// methods implementation, it also takes a "finally" callback which
// is called when the transaction completes, no matter if successfully or not.
type TransactionBase struct {
	mutex   sync.RWMutex
	done    chan struct{}
	err     error
	finally FinallyCallback
}

// NewTransactionBase creates a new TransactionBase.
//
// The finally callback will be called when the transaction finishes, no matter
// if successfully or not. The typical usage is to delete the transaction from
// the transaction store.
func NewTransactionBase(finally func()) *TransactionBase {
	return &TransactionBase{
		done:    make(chan struct{}),
		finally: finally,
	}
}

// Transaction.Done() implementation.
func (t *TransactionBase) Done() <-chan struct{} {
	return t.done
}

// Transaction.Success() implementation.
func (t *TransactionBase) Success() {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.finish()
}

// You must acquire write lock on t.mutex before calling this function!
func (t *TransactionBase) finish() {
	if t.finally != nil {
		t.finally()
	}
	select {
	case <-t.done:
	default:
		close(t.done)
	}
}

// Transaction.Err() implementation.
func (t *TransactionBase) Err() error {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	return t.err
}

// Transaction.Fail() implementation.
func (t *TransactionBase) Fail(e error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.err = e
	t.finish()
}
