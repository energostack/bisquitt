package packets1

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	pkts "github.com/energomonitor/bisquitt/packets"
)

func TestPubackConstructor(t *testing.T) {
	assert := assert.New(t)

	topicID := uint16(123)
	pkt := NewPuback(topicID, RC_ACCEPTED)

	if pkt == nil {
		t.Fatal("New packet should not be nil")
	}

	assert.Equal("*packets1.Puback", reflect.TypeOf(pkt).String(), "Type should be Puback")
	assert.Equal(topicID, pkt.TopicID, fmt.Sprintf("TopicID should be %d", topicID))
	assert.Equal(uint16(0), pkt.MessageID(), "Default MessageID should be 0")
	assert.Equal(RC_ACCEPTED, pkt.ReturnCode, "ReturnCode should be RC_ACCEPTED")
	assert.Equal(uint16(7), pkt.PacketLength(), "Default Length should be 7")
}

func TestPubackMarshal(t *testing.T) {
	pkt1 := NewPuback(123, RC_CONGESTION)
	pkt1.SetMessageID(12)
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*Puback))
}

func TestPubackUnmarshalInvalid(t *testing.T) {
	assert := assert.New(t)

	// Packet too short.
	buff := bytes.NewBuffer([]byte{
		6,                 // Length
		byte(pkts.PUBACK), // MsgType
		0, 1,              // Topic ID
		0, 2, // Message ID
		// Return Code missing
	})
	_, err := ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad PUBACK packet length")
	}

	// Packet too long.
	buff = bytes.NewBuffer([]byte{
		6,                 // Length
		byte(pkts.PUBACK), // MsgType
		0, 1,              // Topic ID
		0, 2, // Message ID
		3, // Return Code
		0, // junk
	})
	_, err = ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad PUBACK packet length")
	}
}

func TestPubackStringer(t *testing.T) {
	pkt := NewPuback(123, RC_CONGESTION)
	pkt.SetMessageID(12)
	assert.Equal(t, "PUBACK(TopicID=123, ReturnCode=1, MessageID=12)", pkt.String())
}
