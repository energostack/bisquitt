package client

import (
	"fmt"

	pkts "github.com/energomonitor/bisquitt/packets1"
	"github.com/energomonitor/bisquitt/transactions"
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
				func(lastMsg interface{}) error {
					tLog.Debug("Resend.")
					return client.send(lastMsg.(pkts.Message))
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

func (t *registerTransaction) Regack(regack *pkts.Regack) {
	if regack.ReturnCode != pkts.RC_ACCEPTED {
		t.Fail(fmt.Errorf("registration rejected with code %d", regack.ReturnCode))
		return
	}

	register := t.Data.(*pkts.Register)
	t.client.registeredTopicsLock.Lock()
	t.client.registeredTopics[register.TopicName] = regack.TopicID
	t.client.registeredTopicsLock.Unlock()
	t.Success()
}
