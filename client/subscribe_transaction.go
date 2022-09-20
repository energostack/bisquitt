package client

import (
	"fmt"
	"strings"

	pkts "github.com/energomonitor/bisquitt/packets"
	pkts1 "github.com/energomonitor/bisquitt/packets1"
	"github.com/energomonitor/bisquitt/transactions"
)

type subscribeTransaction struct {
	*transaction
	callback MessageHandlerFunc
}

func newSubscribeTransaction(client *Client, msgID uint16, callback MessageHandlerFunc) *subscribeTransaction {
	tLog := client.log.WithTag(fmt.Sprintf("SUBSCRIBE(%d)", msgID))
	tLog.Debug("Created.")
	return &subscribeTransaction{
		transaction: &transaction{
			RetryTransaction: transactions.NewRetryTransaction(
				client.groupCtx, client.cfg.RetryDelay, client.cfg.RetryCount,
				func(lastPkt interface{}) error {
					tLog.Debug("Resend.")
					dupPkt := lastPkt.(pkts1.PacketWithDUP)
					dupPkt.SetDUP(true)
					return client.send(lastPkt.(pkts.Packet))
				},
				func() {
					tLog.Debug("Deleted.")
					client.transactions.Delete(msgID)
				},
			),
			client: client,
			log:    tLog,
		},
		callback: callback,
	}
}

func (t *subscribeTransaction) Suback(suback *pkts1.Suback) {
	if suback.ReturnCode != pkts1.RC_ACCEPTED {
		t.Fail(fmt.Errorf("subscription rejected with code %d", suback.ReturnCode))
		return
	}

	var topicName string
	subscribe := t.Data.(*pkts1.Subscribe)

	switch subscribe.TopicIDType {
	case pkts1.TIT_STRING:
		topicName = string(subscribe.TopicName)

		// When subscribing to a wildcard topic, gateway returns TopicID == 0x0000.
		// See `5.4.16 SUBACK` in MQTT-SN 1.2 specification.
		if suback.TopicID == 0 {
			break
		}

		t.log.Debug(`Topic "%s" registered as TopicID %d`,
			topicName,
			suback.TopicID,
		)
		t.client.registeredTopicsLock.Lock()
		t.client.registeredTopics[topicName] = suback.TopicID
		t.client.registeredTopicsLock.Unlock()

	case pkts1.TIT_PREDEFINED:
		var ok bool
		topicName, ok = t.client.cfg.PredefinedTopics.GetTopicName(t.client.cfg.ClientID, subscribe.TopicID)
		if !ok {
			t.Fail(fmt.Errorf("invalid predefined topic ID: %d", subscribe.TopicID))
			return
		}

	case pkts1.TIT_SHORT:
		topicName = pkts1.DecodeShortTopic(subscribe.TopicID)
		break

	default:
		t.Fail(fmt.Errorf("invalid Topic ID Type: %d", subscribe.TopicIDType))
		return
	}

	t.client.messageHandlers.store(
		strings.Split(topicName, "/"),
		t.callback,
	)

	t.Success()
}
