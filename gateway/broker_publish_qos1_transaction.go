package gateway

import (
	"context"
	"fmt"

	mqttPackets "github.com/eclipse/paho.mqtt.golang/packets"
	snPkts "github.com/energomonitor/bisquitt/packets1"
	"github.com/energomonitor/bisquitt/transactions"
)

type brokerPublishQOS1Transaction struct {
	brokerPublishTransactionBase
}

func newBrokerPublishQOS1Transaction(ctx context.Context, h *handler, msgID uint16) *brokerPublishQOS1Transaction {
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

func (t *brokerPublishQOS1Transaction) Regack(snRegack *snPkts.Regack) error {
	return t.regack(snRegack, awaitingPuback)
}

func (t *brokerPublishQOS1Transaction) Puback(snPuback *snPkts.Puback) error {
	if t.State != awaitingPuback {
		t.log.Debug("Unexpected message in %d: %v", t.State, snPuback)
		return nil
	}
	if snPuback.ReturnCode != snPkts.RC_ACCEPTED {
		t.Fail(fmt.Errorf("PUBACK return code: %d", snPuback.ReturnCode))
		return nil
	}
	mqPuback := mqttPackets.NewControlPacket(mqttPackets.Puback).(*mqttPackets.PubackPacket)
	mqPuback.MessageID = snPuback.MessageID()
	return t.ProceedMQTT(transactionDone, mqPuback)
}
