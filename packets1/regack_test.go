package packets1

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	pkts "github.com/energomonitor/bisquitt/packets"
)

func TestRegackConstructor(t *testing.T) {
	assert := assert.New(t)

	topicID := uint16(123)
	returnCode := RC_ACCEPTED
	pkt := NewRegack(topicID, returnCode)

	if pkt == nil {
		t.Fatal("New packet should not be nil")
	}

	assert.Equal("*packets1.Regack", reflect.TypeOf(pkt).String(), "Type should be Regack")
	assert.Equal(uint16(7), pkt.PacketLength(), "Default Length should be 7")
	assert.Equal(topicID, pkt.TopicID, "Bad TopicID value")
	assert.Equal(uint16(0), pkt.MessageID(), "Default MessageID should be 0")
	assert.Equal(returnCode, pkt.ReturnCode, "ReturnCode should be RC_ACCEPTED")
}

func TestRegackMarshal(t *testing.T) {
	pkt1 := NewRegack(1234, RC_CONGESTION)
	pkt1.SetMessageID(2345)
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*Regack))
}

func TestRegackUnmarshalInvalid(t *testing.T) {
	assert := assert.New(t)

	// Packet too short.
	buff := bytes.NewBuffer([]byte{
		6,                 // Length
		byte(pkts.REGACK), // MsgType
		0, 1,              // Topic ID
		0, 2, // Message ID
		// Return Code missing
	})
	_, err := ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad REGACK packet length")
	}

	// Packet too long.
	buff = bytes.NewBuffer([]byte{
		8,                 // Length
		byte(pkts.REGACK), // MsgType
		0, 1,              // Topic ID
		0, 2, // Message ID
		1, // Return Code
		0, // junk
	})
	_, err = ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad REGACK packet length")
	}
}

func TestRegackStringer(t *testing.T) {
	pkt := NewRegack(1234, RC_CONGESTION)
	pkt.SetMessageID(2345)
	assert.Equal(t, "REGACK(TopicID=1234, ReturnCode=1, MessageID=2345)", pkt.String())
}
