package packets1

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	pkts "github.com/energostack/bisquitt/packets"
)

func TestUnsubackConstructor(t *testing.T) {
	assert := assert.New(t)

	pkt := NewUnsuback()

	if pkt == nil {
		t.Fatal("New packet should not be nil")
	}

	assert.Equal("*packets1.Unsuback", reflect.TypeOf(pkt).String(), "Type should be Unsuback")
	assert.Equal(uint16(4), pkt.PacketLength(), "Default Length should be 4")
	assert.Equal(uint16(0), pkt.MessageID(), "Default MessageID should be 0")
}

func TestUnsubackMarshal(t *testing.T) {
	pkt1 := NewUnsuback()
	pkt1.SetMessageID(1234)
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*Unsuback))
}

func TestUnsubackUnmarshalInvalid(t *testing.T) {
	assert := assert.New(t)

	// Packet too short.
	buff := bytes.NewBuffer([]byte{
		3,                   // Length
		byte(pkts.UNSUBACK), // MsgType
		0,                   // Message ID MSB
		// Message ID LSB missing
	})
	_, err := ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad UNSUBACK packet length")
	}

	// Packet too long.
	buff = bytes.NewBuffer([]byte{
		5,                   // Length
		byte(pkts.UNSUBACK), // MsgType
		0, 1,                // Message ID
		0, // junk
	})
	_, err = ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad UNSUBACK packet length")
	}
}

func TestUnsubackStringer(t *testing.T) {
	pkt := NewUnsuback()
	pkt.SetMessageID(1234)
	assert.Equal(t, "UNSUBACK(MessageID=1234)", pkt.String())
}
