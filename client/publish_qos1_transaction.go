package client

import (
	"fmt"

	pkts1 "github.com/energomonitor/bisquitt/packets1"
	"github.com/energomonitor/bisquitt/transactions"
)

type publishQOS1Transaction struct {
	*transaction
}

func newPublishQOS1Transaction(client *Client, msgID uint16) *publishQOS1Transaction {
	tLog := client.log.WithTag(fmt.Sprintf("PUBLISH1(%d)", msgID))
	tLog.Debug("Created.")
	return &publishQOS1Transaction{
		transaction: &transaction{
			RetryTransaction: transactions.NewRetryTransaction(
				client.groupCtx, client.cfg.RetryDelay, client.cfg.RetryCount,
				func(lastPkt interface{}) error {
					tLog.Debug("Resend.")
					return client.send(lastPkt.(pkts1.Packet))
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

func (t *publishQOS1Transaction) Puback(puback *pkts1.Puback) {
	if t.State != awaitingPuback {
		t.log.Debug("Unexpected packet in %d: %v", t.State, puback)
		return
	}
	t.Success()
}
