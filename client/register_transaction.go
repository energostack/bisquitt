package client

import (
	"fmt"

	pkts "github.com/energostack/bisquitt/packets"
	pkts1 "github.com/energostack/bisquitt/packets1"
	"github.com/energostack/bisquitt/transactions"
)

type registerTransaction struct {
	*transaction
}

func newRegisterTransaction(client *Client, msgID uint16, topic string) *registerTransaction {
	tLog := client.log.WithTag(fmt.Sprintf("REGISTER(%d)", msgID))
	tLog.Debug("Created.")
	return &registerTransaction{
		transaction: &transaction{
			RetryTransaction: transactions.NewRetryTransaction(
				client.groupCtx, client.cfg.RetryDelay, client.cfg.RetryCount,
				func(lastPkt interface{}) error {
					tLog.Debug("Resend.")
					return client.send(lastPkt.(pkts.Packet))
				},
				func() {
					client.transactions.Delete(msgID)
					tLog.Debug("Deleted.")
				},
			),
			client: client,
			log:    tLog,
		},
	}
}

func (t *registerTransaction) Regack(regack *pkts1.Regack) {
	if regack.ReturnCode != pkts1.RC_ACCEPTED {
		t.Fail(fmt.Errorf("registration rejected with code %d", regack.ReturnCode))
		return
	}

	register := t.Data.(*pkts1.Register)
	t.client.registeredTopicsLock.Lock()
	t.client.registeredTopics[register.TopicName] = regack.TopicID
	t.client.registeredTopicsLock.Unlock()
	t.Success()
}
