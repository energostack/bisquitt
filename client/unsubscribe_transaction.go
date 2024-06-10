package client

import (
	"fmt"
	"strings"

	pkts "github.com/energostack/bisquitt/packets"
	pkts1 "github.com/energostack/bisquitt/packets1"
	"github.com/energostack/bisquitt/transactions"
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
				func(lastPkt interface{}) error {
					tLog.Debug("Resend.")
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
	}
}

func (t *unsubscribeTransaction) Unsuback(_ *pkts1.Unsuback) {
	var topicName string
	unsubscribe := t.Data.(*pkts1.Unsubscribe)

	switch unsubscribe.TopicIDType {
	case pkts1.TIT_STRING:
		topicName = string(unsubscribe.TopicName)

	case pkts1.TIT_PREDEFINED:
		var ok bool
		topicName, ok = t.client.cfg.PredefinedTopics.GetTopicName(t.client.cfg.ClientID, unsubscribe.TopicID)
		if !ok {
			t.Fail(fmt.Errorf("invalid predefined topic ID: %d", unsubscribe.TopicID))
			return
		}

	case pkts1.TIT_SHORT:
		topicName = pkts.DecodeShortTopic(unsubscribe.TopicID)

	default:
		t.Fail(fmt.Errorf("invalid Topic ID Type: %d", unsubscribe.TopicIDType))
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
