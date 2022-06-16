package gateway

import (
	"context"
	"fmt"

	mqttPackets "github.com/eclipse/paho.mqtt.golang/packets"
	snPkts "github.com/energomonitor/bisquitt/packets1"
	"github.com/energomonitor/bisquitt/transactions"
)

type brokerPublishQOS2Transaction struct {
	brokerPublishTransactionBase
}

func newBrokerPublishQOS2Transaction(ctx context.Context, h *handler, msgID uint16) *brokerPublishQOS2Transaction {
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

func (t *brokerPublishQOS2Transaction) Regack(snRegack *snPkts.Regack) error {
	return t.regack(snRegack, awaitingPubrec)
}

func (t *brokerPublishQOS2Transaction) Pubrec(snPubrec *snPkts.Pubrec) error {
	if t.State != awaitingPubrec {
		t.log.Debug("Unexpected message in %d: %v", t.State, snPubrec)
		return nil
	}
	mqPubrec := mqttPackets.NewControlPacket(mqttPackets.Pubrec).(*mqttPackets.PubrecPacket)
	mqPubrec.MessageID = snPubrec.MessageID()
	return t.ProceedMQTT(awaitingPubrel, mqPubrec)
}

func (t *brokerPublishQOS2Transaction) Pubrel(mqPubrel *mqttPackets.PubrelPacket) error {
	if t.State != awaitingPubrel {
		t.log.Debug("Unexpected message in %d: %v", t.State, mqPubrel)
		return nil
	}
	snPubrel := snPkts.NewPubrel()
	snPubrel.SetMessageID(mqPubrel.MessageID)
	return t.ProceedSN(awaitingPubcomp, snPubrel)
}

func (t *brokerPublishQOS2Transaction) Pubcomp(snPubcomp *snPkts.Pubcomp) error {
	if t.State != awaitingPubcomp {
		t.log.Debug("Unexpected message in %d: %v", t.State, snPubcomp)
		return nil
	}
	mqPubcomp := mqttPackets.NewControlPacket(mqttPackets.Pubcomp).(*mqttPackets.PubcompPacket)
	mqPubcomp.MessageID = snPubcomp.MessageID()
	return t.ProceedMQTT(transactionDone, mqPubcomp)
}
