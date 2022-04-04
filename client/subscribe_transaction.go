package client

import (
	"fmt"
	"strings"

	msgs "github.com/energomonitor/bisquitt/messages"
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
				func(lastMsg interface{}) error {
					tLog.Debug("Resend.")
					dupMsg := lastMsg.(msgs.MessageWithDUP)
					dupMsg.SetDUP(true)
					return client.send(lastMsg.(msgs.Message))
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

func (t *subscribeTransaction) Suback(suback *msgs.SubackMessage) {
	if suback.ReturnCode != msgs.RC_ACCEPTED {
		t.Fail(fmt.Errorf("subscription rejected with code %d", suback.ReturnCode))
		return
	}

	var topicName string
	subscribe := t.Data.(*msgs.SubscribeMessage)

	switch subscribe.TopicIDType {
	case msgs.TIT_STRING:
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

	case msgs.TIT_PREDEFINED:
		var ok bool
		topicName, ok = t.client.cfg.PredefinedTopics.GetTopicName(t.client.cfg.ClientID, subscribe.TopicID)
		if !ok {
			t.Fail(fmt.Errorf("Invalid predefined topic ID: %d", subscribe.TopicID))
			return
		}

	case msgs.TIT_SHORT:
		topicName = msgs.DecodeShortTopic(subscribe.TopicID)
		break

	default:
		t.Fail(fmt.Errorf("Invalid Topic ID Type: %d", subscribe.TopicIDType))
		return
	}

	t.client.messageHandlers.store(
		strings.Split(topicName, "/"),
		t.callback,
	)

	t.Success()
}
