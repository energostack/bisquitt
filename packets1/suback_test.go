package packets1

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	pkts "github.com/energomonitor/bisquitt/packets"
)

func TestSubackConstructor(t *testing.T) {
	assert := assert.New(t)

	topicID := uint16(12)
	qos := uint8(1)
	returnCode := RC_ACCEPTED
	pkt := NewSuback(topicID, returnCode, qos)

	if pkt == nil {
		t.Fatal("New packet should not be nil")
	}

	assert.Equal("*packets1.Suback", reflect.TypeOf(pkt).String(), "Type should be Suback")
	assert.Equal(uint16(8), pkt.PacketLength(), "Default Length should be 8")
	assert.Equal(qos, pkt.QOS, "Bad QOS value")
	assert.Equal(RC_ACCEPTED, pkt.ReturnCode, "ReturnCode should be RC_ACCEPTED")
	assert.Equal(topicID, pkt.TopicID, "Bad TopicID value")
	assert.Equal(uint16(0), pkt.MessageID(), "Default MessageID should be 0")
}

func TestSubackMarshal(t *testing.T) {
	pkt1 := NewSuback(1234, RC_CONGESTION, 1)
	pkt1.SetMessageID(2345)
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*Suback))
}

func TestSubackUnmarshalInvalid(t *testing.T) {
	assert := assert.New(t)

	// Packet too short.
	buff := bytes.NewBuffer([]byte{
		7,                 // Length
		byte(pkts.SUBACK), // MsgType
		0,                 // Flags
		0, 1,              // Topic ID
		0, 2, // Message ID
		// Return Code missing
	})
	_, err := ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad SUBACK packet length")
	}

	// Packet too long.
	buff = bytes.NewBuffer([]byte{
		9,                 // Length
		byte(pkts.SUBACK), // MsgType
		0,                 // Flags
		0, 1,              // Topic ID
		0, 2, // Message ID
		0, // Return Code
		0, // junk
	})
	_, err = ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad SUBACK packet length")
	}
}

func TestSubackStringer(t *testing.T) {
	pkt := NewSuback(1234, RC_CONGESTION, 1)
	pkt.SetMessageID(2345)
	assert.Equal(t, "SUBACK(TopicID=1234, MessageID=2345, ReturnCode=1, QOS=1)", pkt.String())
}
