package packets1

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnsubscribeStruct(t *testing.T) {
	topicID := uint16(12)
	topicIDType := TIT_REGISTERED
	topicName := "test-topic"
	pkt := NewUnsubscribe(topicName, topicID, topicIDType)

	if assert.NotNil(t, pkt, "New packet should not be nil") {
		assert.Equal(t, "*packets1.Unsubscribe", reflect.TypeOf(pkt).String(), "Type should be Unsubscribe")
		assert.Equal(t, topicIDType, pkt.TopicIDType, "Bad TopicIDType value")
		assert.Equal(t, topicID, pkt.TopicID, "Bad TopicID value")
		assert.Equal(t, uint16(0), pkt.MessageID(), "Default MessageID should be 0")
		assert.Equal(t, topicName, pkt.TopicName, "Bad Topicname value")
	}
}

func TestUnsubscribeMarshalString(t *testing.T) {
	pkt1 := NewUnsubscribe("test-topic", 0, TIT_STRING)
	pkt1.SetMessageID(12)
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*Unsubscribe))
}

func TestUnsubscribeMarshalShort(t *testing.T) {
	pkt1 := NewUnsubscribe("", 123, TIT_SHORT)
	pkt1.SetMessageID(12)
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*Unsubscribe))
}
