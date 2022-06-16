// MQTT-SN specification v 1.2 (chapter 5.4.13) says that the PUBACK message contains:
//
// 	TopicId: same value the one contained in the corresponding PUBLISH message.
//
// I don't understand the logic behind this field because:
//
// 1. The PUBACK message contains a MsgId field which already identifies the corresponding
//    PUBLISH message.
// 2. The MQTT PUBACK message of course does not include TopicID
// 3. I.e. the existence of the TopicID field in the PUBACK complicates the implementation
//    of a transparent gateway (it must remember the MsgID -> TopicID mapping)
//
// Therefore I consider the TopicID field unnecessary and would rather not include
// it in the PUBACK message. But we include it for the specification compliance and
// therefore we must implement the clientPublish1Transaction just to remember the
// TopicID value :(

package gateway

import (
	"context"
	"fmt"

	mqttPackets "github.com/eclipse/paho.mqtt.golang/packets"
	snPkts "github.com/energomonitor/bisquitt/packets1"
	"github.com/energomonitor/bisquitt/transactions"
	"github.com/energomonitor/bisquitt/util"
)

type clientPublishQOS1Transaction struct {
	*transactions.TimedTransaction
	handler *handler
	log     util.Logger
	topicID uint16
}

func newClientPublishQOS1Transaction(ctx context.Context, h *handler, msgID uint16, topicID uint16) *clientPublishQOS1Transaction {
	tLog := h.log.WithTag(fmt.Sprintf("PUBLISH1c(%d)", msgID))
	tLog.Debug("Created.")
	return &clientPublishQOS1Transaction{
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

func (t *clientPublishQOS1Transaction) Puback(mqPuback *mqttPackets.PubackPacket) error {
	// MQTT-SN PUBACK message contains ReturnCode field. MQTT PUBACK message
	// does not contain it - PUBACK's implicit meaning is "accepted".
	// See MQTT-SN specification v. 1.2, chapter 5.4.13.
	snPuback := snPkts.NewPuback(t.topicID, snPkts.RC_ACCEPTED)
	snPuback.SetMessageID(mqPuback.MessageID)
	t.Success()
	return t.handler.snSend(snPuback)
}
