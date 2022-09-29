// MQTT-SN SUBACK contains TopicID field. We must use this transaction to save
// MsgID -> TopicID mapping.

package gateway

import (
	"context"
	"fmt"

	mqPkts "github.com/eclipse/paho.mqtt.golang/packets"

	snPkts1 "github.com/energomonitor/bisquitt/packets1"
	"github.com/energomonitor/bisquitt/transactions"
	"github.com/energomonitor/bisquitt/util"
)

type subscribeTransaction struct {
	*transactions.TimedTransaction
	handler *handler1
	log     util.Logger
	topicID uint16
}

func newSubscribeTransaction(ctx context.Context, h *handler1, msgID uint16, topicID uint16) *subscribeTransaction {
	tLog := h.log.WithTag(fmt.Sprintf("REGISTERc(%d)", msgID))
	tLog.Debug("Created.")
	return &subscribeTransaction{
		TimedTransaction: transactions.NewTimedTransaction(
			ctx, h.cfg.RetryDelay,
			func() {
				h.transactions.Delete(msgID)
				tLog.Debug("Deleted.")
			},
		),
		handler: h,
		log:     tLog,
		topicID: topicID,
	}
}

func (t *subscribeTransaction) Suback(mqSuback *mqPkts.SubackPacket) error {
	if len(mqSuback.ReturnCodes) != 1 {
		err := fmt.Errorf("unexpected ReturnCodes length in MQTT/SUBACK: %d", len(mqSuback.ReturnCodes))
		t.Fail(err)
		return err
	}
	// MQTT Return codes 0-2 means "Success, QoS 0-2" but in MQTT-SN only 0
	// means success!
	var returnCode snPkts1.ReturnCode
	if mqSuback.ReturnCodes[0] <= 2 {
		returnCode = snPkts1.RC_ACCEPTED
		t.Success()
	} else {
		returnCode = snPkts1.RC_NOT_SUPPORTED
		t.Fail(fmt.Errorf("MQTT SUBACK return code: %d", mqSuback.ReturnCodes[0]))
	}
	snPkt := snPkts1.NewSuback(t.topicID, returnCode, mqSuback.Qos)
	snPkt.SetMessageID(mqSuback.MessageID)
	return t.handler.snSend(snPkt)
}
