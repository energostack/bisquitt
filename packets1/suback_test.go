package packets1

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSubackStruct(t *testing.T) {
	topicID := uint16(12)
	qos := uint8(1)
	returnCode := RC_ACCEPTED
	pkt := NewSuback(topicID, qos, returnCode)

	if assert.NotNil(t, pkt, "New packet should not be nil") {
		assert.Equal(t, "*packets1.Suback", reflect.TypeOf(pkt).String(), "Type should be Suback")
		assert.Equal(t, uint16(8), pkt.PacketLength(), "Default Length should be 8")
		assert.Equal(t, qos, pkt.QOS, "Bad QOS value")
		assert.Equal(t, RC_ACCEPTED, pkt.ReturnCode, "ReturnCode should be RC_ACCEPTED")
		assert.Equal(t, topicID, pkt.TopicID, "Bad TopicID value")
		assert.Equal(t, uint16(0), pkt.MessageID(), "Default MessageID should be 0")
	}
}

func TestSubackMarshal(t *testing.T) {
	pkt1 := NewSuback(123, 1, RC_CONGESTION)
	pkt1.SetMessageID(12)
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*Suback))
}
