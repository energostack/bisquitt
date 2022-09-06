package packets1

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegisterStruct(t *testing.T) {
	topicID := uint16(123)
	topic := "test-topic"
	pkt := NewRegister(topicID, topic)

	if assert.NotNil(t, pkt, "New packet should not be nil") {
		assert.Equal(t, "*packets1.Register", reflect.TypeOf(pkt).String(), "Type should be Register")
		assert.Equal(t, topicID, pkt.TopicID, "Bad TopicID value")
		assert.Equal(t, uint16(0), pkt.MessageID(), "Default MessageID should be 0")
		assert.Equal(t, topic, pkt.TopicName, "Bad TopicName value")
	}
}

func TestRegisterMarshal(t *testing.T) {
	pkt1 := NewRegister(123, "test-topic")
	pkt1.SetMessageID(12)
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*Register))
}
