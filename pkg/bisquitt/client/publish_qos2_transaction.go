package client

import (
	"fmt"

	msgs "github.com/energomonitor/bisquitt/messages"
	"github.com/energomonitor/bisquitt/transactions"
)

type publishQOS2Transaction struct {
	*transaction
}

func newPublishQOS2Transaction(client *Client, msgID uint16) *publishQOS2Transaction {
	tLog := client.log.WithTag(fmt.Sprintf("PUBLISH2(%d)", msgID))
	tLog.Debug("Created.")
	return &publishQOS2Transaction{
		transaction: &transaction{
			RetryTransaction: transactions.NewRetryTransaction(
				client.groupCtx, client.cfg.RetryDelay, client.cfg.RetryCount,
				func(lastMsg interface{}) error {
					tLog.Debug("Resend.")
					return client.send(lastMsg.(msgs.Message))
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

func (t *publishQOS2Transaction) Pubrec(pubrec *msgs.PubrecMessage) error {
	if t.State != awaitingPubrec {
		t.log.Debug("Unexpected message in %d: %v", t.State, pubrec)
		return nil
	}
	pubrel := msgs.NewPubrelMessage()
	pubrel.CopyMessageID(pubrec)
	t.Proceed(awaitingPubcomp, pubrel)
	if err := t.client.send(pubrel); err != nil {
		return err
	}
	return nil
}

func (t *publishQOS2Transaction) Pubcomp(pubcomp *msgs.PubcompMessage) {
	if t.State != awaitingPubcomp {
		t.log.Debug("Unexpected message in %d: %v", t.State, pubcomp)
		return
	}
	t.Success()
}
