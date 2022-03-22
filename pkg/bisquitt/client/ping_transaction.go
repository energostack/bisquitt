package client

import (
	msgs "github.com/energomonitor/bisquitt/messages"
	"github.com/energomonitor/bisquitt/transactions"
)

type pingTransaction struct {
	*transaction
}

func newPingTransaction(client *Client) *pingTransaction {
	tLog := client.log.WithTag("PING")
	tLog.Debug("Created.")
	return &pingTransaction{
		transaction: &transaction{
			RetryTransaction: transactions.NewRetryTransaction(
				client.groupCtx, client.cfg.RetryDelay, client.cfg.RetryCount,
				func(lastMsg interface{}) error {
					tLog.Debug("Resend.")
					return client.send(lastMsg.(msgs.Message))
				},
				func() {
					client.transactions.DeleteByType(msgs.PINGREQ)
					tLog.Debug("Deleted.")
				},
			),
			client: client,
			log:    tLog,
		},
	}
}

func (t *pingTransaction) Pingresp(pingresp *msgs.PingrespMessage) {
	t.Success()
}
