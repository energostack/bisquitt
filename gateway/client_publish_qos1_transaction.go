// MQTT-SN specification v 1.2 (chapter 5.4.13) says that the PUBACK packet contains:
//
// 	TopicId: same value the one contained in the corresponding PUBLISH packet.
//
// I don't understand the logic behind this field because:
//
// 1. The PUBACK packet contains a MsgId field which already identifies the corresponding
//    PUBLISH packet.
// 2. The MQTT PUBACK packet of course does not include TopicID
// 3. I.e. the existence of the TopicID field in the PUBACK complicates the implementation
//    of a transparent gateway (it must remember the MsgID -> TopicID mapping)
//
// Therefore I consider the TopicID field unnecessary and would rather not include
// it in the PUBACK packet. But we include it for the specification compliance and
// therefore we must implement the clientPublish1Transaction just to remember the
// TopicID value :(

package gateway

import (
	"context"
	"fmt"

	mqPkts "github.com/eclipse/paho.mqtt.golang/packets"

	snPkts1 "github.com/energostack/bisquitt/packets1"
	"github.com/energostack/bisquitt/transactions"
	"github.com/energostack/bisquitt/util"
)

type clientPublishQOS1Transaction struct {
	*transactions.TimedTransaction
	handler *handler1
	log     util.Logger
	topicID uint16
}

func newClientPublishQOS1Transaction(ctx context.Context, h *handler1, msgID uint16, topicID uint16) *clientPublishQOS1Transaction {
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

func (t *clientPublishQOS1Transaction) Puback(mqPuback *mqPkts.PubackPacket) error {
	// MQTT-SN PUBACK packet contains ReturnCode field. MQTT PUBACK packet
	// does not contain it - PUBACK's implicit meaning is "accepted".
	// See MQTT-SN specification v. 1.2, chapter 5.4.13.
	snPuback := snPkts1.NewPuback(t.topicID, snPkts1.RC_ACCEPTED)
	snPuback.SetMessageID(mqPuback.MessageID)
	t.Success()
	return t.handler.snSend(snPuback)
}
