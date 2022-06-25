package packets1

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPubackStruct(t *testing.T) {
	topicID := uint16(123)
	pkt := NewPuback(topicID, RC_ACCEPTED)

	if assert.NotNil(t, pkt, "New packet should not be nil") {
		assert.Equal(t, "*packets1.Puback", reflect.TypeOf(pkt).String(), "Type should be Puback")
		assert.Equal(t, topicID, pkt.TopicID, fmt.Sprintf("TopicID should be %d", topicID))
		assert.Equal(t, uint16(0), pkt.MessageID(), "Default MessageID should be 0")
		assert.Equal(t, RC_ACCEPTED, pkt.ReturnCode, "ReturnCode should be RC_ACCEPTED")
		assert.Equal(t, uint16(7), pkt.PacketLength(), "Default Length should be 2")
	}
}

func TestPubackMarshal(t *testing.T) {
	assert := assert.New(t)
	buf := bytes.NewBuffer(nil)

	pkt1 := NewPuback(123, RC_CONGESTION)
	pkt1.SetMessageID(12)
	if err := pkt1.Write(buf); err != nil {
		t.Fatal(err)
	}

	r := bytes.NewReader(buf.Bytes())
	pkt2, err := ReadPacket(r)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(pkt1, pkt2.(*Puback))
}
