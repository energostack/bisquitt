package packets1

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPublishStruct(t *testing.T) {
	dup := true
	retain := true
	qos := uint8(1)
	topicIDType := TIT_SHORT
	topicID := uint16(123)
	payload := []byte("test-payload")
	pkt := NewPublish(topicID, topicIDType, payload, qos, retain, dup)

	if assert.NotNil(t, pkt, "New packet should not be nil") {
		assert.Equal(t, "*packets1.Publish", reflect.TypeOf(pkt).String(), "Type should be Publish")
		assert.Equal(t, dup, pkt.DUP(), "Bad Dup flag value")
		assert.Equal(t, retain, pkt.Retain, "Bad Retain flag value")
		assert.Equal(t, qos, pkt.QOS, "Bad QOS value")
		assert.Equal(t, topicIDType, pkt.TopicIDType, "Bad TopicIDType value")
		assert.Equal(t, topicID, pkt.TopicID, "Bad TopicID value")
		assert.Equal(t, uint16(0), pkt.MessageID(), "Default MessageID should be 0")
		assert.Equal(t, payload, pkt.Data, "Bad Data value")
	}
}

func TestPublishMarshal(t *testing.T) {
	pkt1 := NewPublish(123, TIT_PREDEFINED,
		[]byte("test-payload"), 1, true, true)
	pkt1.SetMessageID(12)
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*Publish))
}
