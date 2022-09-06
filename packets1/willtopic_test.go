package packets1

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWillTopicStruct(t *testing.T) {
	qos := uint8(1)
	retain := false
	willTopic := "test-topic"
	pkt := NewWillTopic(willTopic, qos, retain)

	if assert.NotNil(t, pkt, "New packet should not be nil") {
		assert.Equal(t, "*packets1.WillTopic", reflect.TypeOf(pkt).String(), "Type should be WillTopic")
		assert.Equal(t, qos, pkt.QOS, "Bad QOS value")
		assert.Equal(t, retain, pkt.Retain, "Bad Retain flag value")
		assert.Equal(t, willTopic, pkt.WillTopic, "Bad WillTopic value")
	}
}

func TestWillTopicMarshal(t *testing.T) {
	pkt1 := NewWillTopic("test-topic", 1, true)
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*WillTopic))
}
