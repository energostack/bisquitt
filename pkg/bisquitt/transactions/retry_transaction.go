package transactions

import (
	"context"
	"errors"
	"sync"
	"time"
)

// RetryTransaction implements a StatefulTransaction with automatic time-based
// limited-repetition retries for every transaction step.
//
// E.g. a PUBLISH message should be sent N_retry times with T_retry delay unless
// correspondent PUBACK is received.
type RetryTransaction struct {
	*TransactionBase
	retryDelay    time.Duration
	retryCount    uint
	retryNumMutex sync.Mutex
	retryNum      uint
	timer         *time.Timer
	retryCallback RTRetryCallback
	State         interface{}
	Data          interface{}
}

// Retry callback type.
type RTRetryCallback func(msg interface{}) error

// ErrNoMoreRetries error signalizes that the retry callback was called retryCount
// times in succession and another retryDelay passed without Proceed, Success nor
// Fail being called.
var ErrNoMoreRetries = errors.New("no more retries")

// NewRetryTransaction creates a new RetryTransaction.
//
// In each transaction step, if retryDelay time passes without Proceed, Success,
// or Fail being called, retryCallback is called. If retryCallback is called
// retryCount times in succession and another retryDelay passes without Proceed,
// Success nor Fail being called, the transaction fails.
func NewRetryTransaction(ctx context.Context, retryDelay time.Duration, retryCount uint, retryCallback RTRetryCallback, finally FinallyCallback) *RetryTransaction {
	t := &RetryTransaction{
		TransactionBase: NewTransactionBase(finally),
		retryDelay:      retryDelay,
		retryCount:      retryCount,
		retryCallback:   retryCallback,
	}
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
func (t *RetryTransaction) Success() {
	t.stopTimer()
	t.TransactionBase.Success()
}

// Transaction.Fail() implementation.
func (t *RetryTransaction) Fail(e error) {
	t.stopTimer()
	t.TransactionBase.Fail(e)
}

// StatefulTransaction.Proceed() implementation.
func (t *RetryTransaction) Proceed(state interface{}, data interface{}) {
	t.retryNumMutex.Lock()
	defer t.retryNumMutex.Unlock()

	t.State = state
	t.Data = data
	t.retryNum = 0
	t.restartTimer()
}

func (t *RetryTransaction) stopTimer() {
	if t.timer != nil {
		t.timer.Stop()
	}
}

func (t *RetryTransaction) restartTimer() {
	t.stopTimer()
	t.timer = time.AfterFunc(t.retryDelay, t.timeout)
}

func (t *RetryTransaction) timeout() {
	t.retryNumMutex.Lock()
	defer t.retryNumMutex.Unlock()

	t.retryNum++
	if t.retryNum > t.retryCount {
		t.Fail(ErrNoMoreRetries)
		return
	}
	if err := t.retryCallback(t.Data); err != nil {
		t.Fail(err)
	}
	t.restartTimer()
}
