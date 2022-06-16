package client

import (
	"fmt"
	"time"

	pkts "github.com/energomonitor/bisquitt/packets1"
	"github.com/energomonitor/bisquitt/transactions"
	"github.com/energomonitor/bisquitt/util"
)

const maxPingrespWait = time.Minute

type sleepTransaction struct {
	*transactions.TransactionBase
	client              *Client
	log                 util.Logger
	disconnect          *pkts.DisconnectMessage
	retryDelay          time.Duration
	retryCount          uint
	disconnectResendNum uint
	sleepDuration       time.Duration
	state               transactionState
	timer               *time.Timer
}

func newSleepTransaction(client *Client, sleepDuration time.Duration) *sleepTransaction {
	tLog := client.log.WithTag("SLEEP")
	tLog.Debug("Created.")
	t := &sleepTransaction{
		client:        client,
		log:           tLog,
		retryDelay:    client.cfg.RetryDelay,
		retryCount:    client.cfg.RetryCount,
		sleepDuration: sleepDuration,
	}
	t.TransactionBase = transactions.NewTransactionBase(
		func() {
			t.stopTimer()
			client.transactions.DeleteByType(pkts.DISCONNECT)
			tLog.Debug("Deleted.")
		})

	client.group.Go(func() error {
		select {
		case <-client.groupCtx.Done():
		case <-t.Done():
		}
		return nil
	})
	return t
}

func (t *sleepTransaction) Success() {
	t.stopTimer()
	t.TransactionBase.Success()
}

func (t *sleepTransaction) Fail(e error) {
	t.stopTimer()
	t.TransactionBase.Fail(e)
}

func (t *sleepTransaction) Sleep() error {
	state := t.client.state.Get()
	switch state {
	case util.StateActive:
		duration := uint16(t.sleepDuration / time.Second)
		t.disconnect = pkts.NewDisconnectMessage(duration)
		t.state = awaitingDisconnect
		if err := t.client.send(t.disconnect); err != nil {
			t.Fail(err)
			return err
		}
		t.timer = time.AfterFunc(t.retryDelay, t.resendDisconnect)
	case util.StateAwake:
		t.startSleep()
	default:
		return fmt.Errorf("cannot call Sleep() in %q state", state)
	}
	return nil
}

func (t *sleepTransaction) resendDisconnect() {
	t.disconnectResendNum++
	if t.disconnectResendNum > t.retryCount {
		t.log.Debug("DISCONNECT reply timeout.")
		t.Fail(transactions.ErrNoMoreRetries)
		return
	}
	t.log.Debug("DISCONNECT resend no. %d", t.disconnectResendNum)
	if err := t.client.send(t.disconnect); err != nil {
		t.Fail(err)
		return
	}
	t.timer = time.AfterFunc(t.retryDelay, t.resendDisconnect)
}

func (t *sleepTransaction) Disconnect(disconnect *pkts.DisconnectMessage) {
	if t.state != awaitingDisconnect {
		t.log.Debug("Unexpected message in %d: %v", t.state, disconnect)
		return
	}
	t.stopTimer()
	t.disconnect = nil
	t.startSleep()
}

func (t *sleepTransaction) Pingresp(pingresp *pkts.PingrespMessage) {
	if t.state != awaitingPingresp {
		t.log.Debug("Unexpected message in %d: %v", t.state, pingresp)
		return
	}
	t.stopTimer()
	t.Success()
}

func (t *sleepTransaction) stopTimer() {
	if t.timer != nil {
		t.timer.Stop()
	}
}

func (t *sleepTransaction) startSleep() {
	t.log.Debug("Sleeping for %v...", t.sleepDuration)
	t.client.setState(util.StateAsleep)
	t.timer = time.AfterFunc(t.sleepDuration, t.wakeup)
}

func (t *sleepTransaction) wakeup() {
	t.client.setState(util.StateAwake)
	t.log.Debug("Awake")
	t.state = awaitingPingresp
	ping := pkts.NewPingreqMessage([]byte(t.client.cfg.ClientID))
	if err := t.client.send(ping); err != nil {
		t.Fail(err)
		return
	}
	t.timer = time.AfterFunc(maxPingrespWait, func() {
		t.Fail(fmt.Errorf("did not receive PINGRESP in %v", maxPingrespWait))
	})
}
