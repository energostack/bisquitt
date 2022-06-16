package client

import (
	"fmt"
	"strings"

	pkts "github.com/energomonitor/bisquitt/packets1"
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
					return client.send(lastMsg.(pkts.Message))
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

func (t *unsubscribeTransaction) Unsuback(_ *pkts.UnsubackMessage) {
	var topicName string
	unsubscribe := t.Data.(*pkts.UnsubscribeMessage)

	switch unsubscribe.TopicIDType {
	case pkts.TIT_STRING:
		topicName = string(unsubscribe.TopicName)

	case pkts.TIT_PREDEFINED:
		var ok bool
		topicName, ok = t.client.cfg.PredefinedTopics.GetTopicName(t.client.cfg.ClientID, unsubscribe.TopicID)
		if !ok {
			t.Fail(fmt.Errorf("Invalid predefined topic ID: %d", unsubscribe.TopicID))
			return
		}

	case pkts.TIT_SHORT:
		topicName = pkts.DecodeShortTopic(unsubscribe.TopicID)

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
