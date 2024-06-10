package gateway

import (
	"context"
	"fmt"

	mqPkts "github.com/eclipse/paho.mqtt.golang/packets"

	snPkts1 "github.com/energostack/bisquitt/packets1"
	"github.com/energostack/bisquitt/transactions"
)

type brokerPublishQOS2Transaction struct {
	brokerPublishTransactionBase
}

func newBrokerPublishQOS2Transaction(ctx context.Context, h *handler1, msgID uint16) *brokerPublishQOS2Transaction {
	tLog := h.log.WithTag(fmt.Sprintf("PUBLISH2(%d)", msgID))
	tLog.Debug("Created.")
	t := &brokerPublishQOS2Transaction{
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

func (t *brokerPublishQOS2Transaction) Regack(snRegack *snPkts1.Regack) error {
	return t.regack(snRegack, awaitingPubrec)
}

func (t *brokerPublishQOS2Transaction) Pubrec(snPubrec *snPkts1.Pubrec) error {
	if t.State != awaitingPubrec {
		t.log.Debug("Unexpected packet in %d: %v", t.State, snPubrec)
		return nil
	}
	mqPubrec := mqPkts.NewControlPacket(mqPkts.Pubrec).(*mqPkts.PubrecPacket)
	mqPubrec.MessageID = snPubrec.MessageID()
	return t.ProceedMQTT(awaitingPubrel, mqPubrec)
}

func (t *brokerPublishQOS2Transaction) Pubrel(mqPubrel *mqPkts.PubrelPacket) error {
	if t.State != awaitingPubrel {
		t.log.Debug("Unexpected packet in %d: %v", t.State, mqPubrel)
		return nil
	}
	snPubrel := snPkts1.NewPubrel()
	snPubrel.SetMessageID(mqPubrel.MessageID)
	return t.ProceedSN(awaitingPubcomp, snPubrel)
}

func (t *brokerPublishQOS2Transaction) Pubcomp(snPubcomp *snPkts1.Pubcomp) error {
	if t.State != awaitingPubcomp {
		t.log.Debug("Unexpected packet in %d: %v", t.State, snPubcomp)
		return nil
	}
	mqPubcomp := mqPkts.NewControlPacket(mqPkts.Pubcomp).(*mqPkts.PubcompPacket)
	mqPubcomp.MessageID = snPubcomp.MessageID()
	return t.ProceedMQTT(transactionDone, mqPubcomp)
}
