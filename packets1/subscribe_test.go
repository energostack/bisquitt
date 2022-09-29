package packets1

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSubscribeStruct(t *testing.T) {
	topicID := uint16(12)
	topicIDType := TIT_REGISTERED
	topicName := "test-topic"
	qos := uint8(1)
	dup := true
	pkt := NewSubscribe(topicName, topicID, dup, qos, topicIDType)

	if assert.NotNil(t, pkt, "New packet should not be nil") {
		assert.Equal(t, "*packets1.Subscribe", reflect.TypeOf(pkt).String(), "Type should be Subscribe")
		assert.Equal(t, dup, pkt.DUP(), "Bad Dup flag value")
		assert.Equal(t, qos, pkt.QOS, "Bad QOS value")
		assert.Equal(t, topicIDType, pkt.TopicIDType, "Bad TopicIDType value")
		assert.Equal(t, topicID, pkt.TopicID, "Bad TopicID value")
		assert.Equal(t, uint16(0), pkt.MessageID(), "Default MessageID should be 0")
		assert.Equal(t, topicName, pkt.TopicName, "Bad Topicname value")
	}
}

func TestSubscribeMarshalString(t *testing.T) {
	pkt1 := NewSubscribe("test-topic", 0, true, 1, TIT_STRING)
	pkt1.SetMessageID(12)
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*Subscribe))
}

func TestSubscribeMarshalShort(t *testing.T) {
	pkt1 := NewSubscribe("", 123, true, 1, TIT_SHORT)
	pkt1.SetMessageID(12)
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*Subscribe))
}
