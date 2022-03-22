package transactions

import (
	"context"
	"errors"
	"time"
)

// ErrTimeout error signalizes that a timeout has passed without Success nor Fail
// being called.
var ErrTimeout = errors.New("transaction timeout")

// TimedTransaction fails if Success is not called before the given timeout.
type TimedTransaction struct {
	*TransactionBase
	timer *time.Timer
}

// NewTimedTransaction creates a new TimedTransaction.
func NewTimedTransaction(ctx context.Context, timeout time.Duration, finally FinallyCallback) *TimedTransaction {
	t := &TimedTransaction{
		TransactionBase: NewTransactionBase(finally),
	}
	t.timer = time.AfterFunc(timeout, func() { t.Fail(ErrTimeout) })
	go func() {
		select {
		case <-ctx.Done():
			t.stopTimer()
		case <-t.Done():
			return
		}
	}()
	return t
}

// Transaction.Success() implementation.
func (t *TimedTransaction) Success() {
	t.stopTimer()
	t.TransactionBase.Success()
}

// Transaction.Fail() implementation.
func (t *TimedTransaction) Fail(e error) {
	t.stopTimer()
	t.TransactionBase.Fail(e)
}

func (t *TimedTransaction) stopTimer() {
	t.timer.Stop()
}
