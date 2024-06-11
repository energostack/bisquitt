package gateway

import (
	"context"
	"fmt"

	mqPkts "github.com/eclipse/paho.mqtt.golang/packets"

	snPkts1 "github.com/energostack/bisquitt/packets1"
	"github.com/energostack/bisquitt/transactions"
)

type brokerPublishQOS1Transaction struct {
	brokerPublishTransactionBase
}

func newBrokerPublishQOS1Transaction(ctx context.Context, h *handler1, msgID uint16) *brokerPublishQOS1Transaction {
	tLog := h.log.WithTag(fmt.Sprintf("PUBLISH1(%d)", msgID))
	tLog.Debug("Created.")
	t := &brokerPublishQOS1Transaction{
		brokerPublishTransactionBase: brokerPublishTransactionBase{
			log:     tLog,
			handler: h,
		},
	}
	t.RetryTransaction = transactions.NewRetryTransaction(
		ctx, h.cfg.RetryDelay, h.cfg.RetryCount, t.resend,
		func() {
			h.transactions.Delete(msgID)
			tLog.Debug("Deleted.")
		},
	)
	return t
}

func (t *brokerPublishQOS1Transaction) Regack(snRegack *snPkts1.Regack) error {
	return t.regack(snRegack, awaitingPuback)
}

func (t *brokerPublishQOS1Transaction) Puback(snPuback *snPkts1.Puback) error {
	if t.State != awaitingPuback {
		t.log.Debug("Unexpected packet in %d: %v", t.State, snPuback)
		return nil
	}
	if snPuback.ReturnCode != snPkts1.RC_ACCEPTED {
		t.Fail(fmt.Errorf("PUBACK return code: %d", snPuback.ReturnCode))
		return nil
	}
	mqPuback := mqPkts.NewControlPacket(mqPkts.Puback).(*mqPkts.PubackPacket)
	mqPuback.MessageID = snPuback.MessageID()
	return t.ProceedMQTT(transactionDone, mqPuback)
}
