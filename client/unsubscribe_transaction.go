package client

import (
	"fmt"
	"strings"

	msgs "github.com/energomonitor/bisquitt/messages"
	"github.com/energomonitor/bisquitt/transactions"
)

type unsubscribeTransaction struct {
	*transaction
}

func newUnsubscribeTransaction(client *Client, msgID uint16) *unsubscribeTransaction {
	tLog := client.log.WithTag(fmt.Sprintf("UNSUBSCRIBE(%d)", msgID))
	tLog.Debug("Created.")
	return &unsubscribeTransaction{
		transaction: &transaction{
			RetryTransaction: transactions.NewRetryTransaction(
				client.groupCtx, client.cfg.RetryDelay, client.cfg.RetryCount,
				func(lastMsg interface{}) error {
					tLog.Debug("Resend.")
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
	}
}

func (t *unsubscribeTransaction) Unsuback(_ *msgs.UnsubackMessage) {
	var topicName string
	unsubscribe := t.Data.(*msgs.UnsubscribeMessage)

	switch unsubscribe.TopicIDType {
	case msgs.TIT_STRING:
		topicName = string(unsubscribe.TopicName)

	case msgs.TIT_PREDEFINED:
		var ok bool
		topicName, ok = t.client.cfg.PredefinedTopics.GetTopicName(t.client.cfg.ClientID, unsubscribe.TopicID)
		if !ok {
			t.Fail(fmt.Errorf("Invalid predefined topic ID: %d", unsubscribe.TopicID))
			return
		}

	case msgs.TIT_SHORT:
		topicName = msgs.DecodeShortTopic(unsubscribe.TopicID)

	default:
		t.Fail(fmt.Errorf("Invalid Topic ID Type: %d", unsubscribe.TopicIDType))
		return
	}

	t.log.Debug(`Topic "%s" unsubscribed`,
		topicName,
	)

	t.client.messageHandlers.delete(
		strings.Split(topicName, "/"),
	)

	t.Success()
}
