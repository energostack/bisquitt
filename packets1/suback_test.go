package packets1

import (
	"bytes"
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
	assert := assert.New(t)
	buf := bytes.NewBuffer(nil)

	pkt1 := NewSuback(123, 1, RC_CONGESTION)
	pkt1.SetMessageID(12)
	if err := pkt1.Write(buf); err != nil {
		t.Fatal(err)
	}

	r := bytes.NewReader(buf.Bytes())
	pkt2, err := ReadPacket(r)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(pkt1, pkt2.(*Suback))
}
