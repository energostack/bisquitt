package packets1

import (
	"bytes"
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
	assert := assert.New(t)
	buf := bytes.NewBuffer(nil)

	pkt1 := NewRegack(123, RC_CONGESTION)
	pkt1.SetMessageID(12)
	if err := pkt1.Write(buf); err != nil {
		t.Fatal(err)
	}

	r := bytes.NewReader(buf.Bytes())
	pkt2, err := ReadPacket(r)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(pkt1, pkt2.(*Regack))
}
