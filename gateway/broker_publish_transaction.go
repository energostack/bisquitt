package gateway

import (
	"fmt"

	mqttPackets "github.com/eclipse/paho.mqtt.golang/packets"
	snPkts "github.com/energomonitor/bisquitt/packets1"
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
	Regack(snRegack *snPkts.Regack) error
}

type brokerPublishTransaction interface {
	transactions.StatefulTransaction
	SetSNPublish(*snPkts.Publish)
	ProceedSN(newState transactionState, snMsg snPkts.Packet) error
	ProceedMQTT(newState transactionState, mqMsg mqttPackets.ControlPacket) error
}

type brokerPublishTransactionBase struct {
	*transactions.RetryTransaction
	log       util.Logger
	snPublish *snPkts.Publish
	handler   *handler
}

func (t *brokerPublishTransactionBase) SetSNPublish(snPublish *snPkts.Publish) {
	t.snPublish = snPublish
}

func (t *brokerPublishTransactionBase) regack(snRegack *snPkts.Regack, newState transactionState) error {
	if t.State != awaitingRegack {
		t.log.Debug("Unexpected message in %d: %v", t.State, snRegack)
		return nil
	}
	if snRegack.ReturnCode != snPkts.RC_ACCEPTED {
		t.Fail(fmt.Errorf("REGACK return code: %d", snRegack.ReturnCode))
		return nil
	}
	snRegister := t.Data.(*snPkts.Register)
	t.handler.registeredTopics.Store(snRegister.TopicID, snRegister.TopicName)
	return t.ProceedSN(newState, t.snPublish)
}

func (t *brokerPublishTransactionBase) ProceedSN(newState transactionState, snMsg snPkts.Packet) error {
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

func (t *brokerPublishTransactionBase) ProceedMQTT(newState transactionState, mqMsg mqttPackets.ControlPacket) error {
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

// Resend MQTT or MQTT-SN message.
func (t *brokerPublishTransactionBase) resend(msgx interface{}) error {
	t.log.Debug("Resend.")
	switch msg := msgx.(type) {
	case snPkts.Packet:
		// Set DUP if applicable.
		if dupMsg, ok := msg.(snPkts.PacketWithDUP); ok {
			dupMsg.SetDUP(true)
		}
		return t.handler.snSend(msg)
	case mqttPackets.ControlPacket:
		// PUBLISH is the only message with DUP in MQTT.
		if publish, ok := msg.(*mqttPackets.PublishPacket); ok {
			publish.Dup = true
		}
		return t.handler.mqttSend(msg)
	default:
		return fmt.Errorf("invalid message type (%T): %v", msgx, msgx)
	}
}
