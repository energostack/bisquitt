package client

import (
	"fmt"

	pkts1 "github.com/energomonitor/bisquitt/packets1"
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

func (t *publishQOS2Transaction) Pubrec(pubrec *pkts1.Pubrec) error {
	if t.State != awaitingPubrec {
		t.log.Debug("Unexpected packet in %d: %v", t.State, pubrec)
		return nil
	}
	pubrel := pkts1.NewPubrel()
	pubrel.CopyMessageID(pubrec)
	t.Proceed(awaitingPubcomp, pubrel)
	if err := t.client.send(pubrel); err != nil {
		return err
	}
	return nil
}

func (t *publishQOS2Transaction) Pubcomp(pubcomp *pkts1.Pubcomp) {
	if t.State != awaitingPubcomp {
		t.log.Debug("Unexpected packet in %d: %v", t.State, pubcomp)
		return
	}
	t.Success()
}
