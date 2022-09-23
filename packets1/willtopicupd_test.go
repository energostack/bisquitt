package packets1

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWillTopicUpdStruct(t *testing.T) {
	willTopic := "test-topic"
	qos := uint8(1)
	retain := true
	pkt := NewWillTopicUpd(willTopic, qos, retain)

	if assert.NotNil(t, pkt, "New packet should not be nil") {
		assert.Equal(t, "*packets1.WillTopicUpd", reflect.TypeOf(pkt).String(), "Type should be WillTopicUpd")
		assert.Equal(t, qos, pkt.QOS, "Bad QOS value")
		assert.Equal(t, retain, pkt.Retain, "Bad Retain flag value")
		assert.Equal(t, willTopic, pkt.WillTopic, "Bad WillTopic value")
	}

}

func TestWillTopicUpdMarshal(t *testing.T) {
	pkt1 := NewWillTopicUpd("test-topic", 1, true)
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*WillTopicUpd))
}
