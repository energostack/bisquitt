package gateway

import (
	"fmt"

	mqPkts "github.com/eclipse/paho.mqtt.golang/packets"

	snPkts1 "github.com/energomonitor/bisquitt/packets1"
	"github.com/energomonitor/bisquitt/transactions"
	"github.com/energomonitor/bisquitt/util"
)

// Transactions states constants
type transactionState int

const (
	transactionDone transactionState = iota
	awaitingRegack
	awaitingPuback
	awaitingPubrec
	awaitingPubrel
	awaitingPubcomp
)

type transactionWithRegack interface {
	Regack(snRegack *snPkts1.Regack) error
}

type brokerPublishTransaction interface {
	transactions.StatefulTransaction
	SetSNPublish(*snPkts1.Publish)
	ProceedSN(newState transactionState, snMsg snPkts1.Packet) error
	ProceedMQTT(newState transactionState, mqMsg mqPkts.ControlPacket) error
}

type brokerPublishTransactionBase struct {
	*transactions.RetryTransaction
	log       util.Logger
	snPublish *snPkts1.Publish
	handler   *handler1
}

func (t *brokerPublishTransactionBase) SetSNPublish(snPublish *snPkts1.Publish) {
	t.snPublish = snPublish
}

func (t *brokerPublishTransactionBase) regack(snRegack *snPkts1.Regack, newState transactionState) error {
	if t.State != awaitingRegack {
		t.log.Debug("Unexpected packet in %d: %v", t.State, snRegack)
		return nil
	}
	if snRegack.ReturnCode != snPkts1.RC_ACCEPTED {
		t.Fail(fmt.Errorf("REGACK return code: %d", snRegack.ReturnCode))
		return nil
	}
	snRegister := t.Data.(*snPkts1.Register)
	t.handler.registeredTopics.Store(snRegister.TopicID, snRegister.TopicName)
	return t.ProceedSN(newState, t.snPublish)
}

func (t *brokerPublishTransactionBase) ProceedSN(newState transactionState, snMsg snPkts1.Packet) error {
	t.Proceed(newState, snMsg)
	if err := t.handler.snSend(snMsg); err != nil {
		t.Fail(err)
		return err
	}
	if newState == transactionDone {
		t.Success()
	}
	return nil
}

func (t *brokerPublishTransactionBase) ProceedMQTT(newState transactionState, mqMsg mqPkts.ControlPacket) error {
	t.Proceed(newState, mqMsg)
	if err := t.handler.mqttSend(mqMsg); err != nil {
		t.Fail(err)
		return err
	}
	if newState == transactionDone {
		t.Success()
	}
	return nil
}

// Resend MQTT or MQTT-SN packet.
func (t *brokerPublishTransactionBase) resend(pktx interface{}) error {
	t.log.Debug("Resend.")
	switch pkt := pktx.(type) {
	case snPkts1.Packet:
		// Set DUP if applicable.
		if dupMsg, ok := pkt.(snPkts1.PacketWithDUP); ok {
			dupMsg.SetDUP(true)
		}
		return t.handler.snSend(pkt)
	case mqPkts.ControlPacket:
		// PUBLISH is the only packet with DUP in MQTT.
		if publish, ok := pkt.(*mqPkts.PublishPacket); ok {
			publish.Dup = true
		}
		return t.handler.mqttSend(pkt)
	default:
		return fmt.Errorf("invalid package type (%T): %v", pktx, pktx)
	}
}
