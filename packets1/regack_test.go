package packets1

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegackStruct(t *testing.T) {
	topicID := uint16(123)
	returnCode := RC_ACCEPTED
	pkt := NewRegack(topicID, returnCode)

	if assert.NotNil(t, pkt, "New packet should not be nil") {
		assert.Equal(t, "*packets1.Regack", reflect.TypeOf(pkt).String(), "Type should be Regack")
		assert.Equal(t, uint16(7), pkt.PacketLength(), "Default Length should be 7")
		assert.Equal(t, topicID, pkt.TopicID, "Bad TopicID value")
		assert.Equal(t, uint16(0), pkt.MessageID(), "Default MessageID should be 0")
		assert.Equal(t, returnCode, pkt.ReturnCode, "ReturnCode should be RC_ACCEPTED")
	}
}

func TestRegackMarshal(t *testing.T) {
	pkt1 := NewRegack(123, RC_CONGESTION)
	pkt1.SetMessageID(12)
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*Regack))
}
