package client

import (
	pkts "github.com/energomonitor/bisquitt/packets"
	pkts1 "github.com/energomonitor/bisquitt/packets1"
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
				func(lastPkt interface{}) error {
					tLog.Debug("Resend.")
					return client.send(lastPkt.(pkts1.Packet))
				},
				func() {
					client.transactions.DeleteByType(pkts.PINGREQ)
					tLog.Debug("Deleted.")
				},
			),
			client: client,
			log:    tLog,
		},
	}
}

func (t *pingTransaction) Pingresp(pingresp *pkts1.Pingresp) {
	t.Success()
}
