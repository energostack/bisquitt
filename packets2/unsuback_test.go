package packets2

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	pkts "github.com/energomonitor/bisquitt/packets"
)

func TestUnsubackConstructor(t *testing.T) {
	assert := assert.New(t)

	pkt := NewUnsuback(RC_CONGESTION)

	if pkt == nil {
		t.Fatal("New packet should not be nil")
	}

	assert.Equal("*packets2.Unsuback", reflect.TypeOf(pkt).String(), "Type should be Unsuback")
	assert.Equal(RC_CONGESTION, pkt.ReasonCode, "Bad ReasonCode value")
	assert.Equal(uint16(5), pkt.PacketLength(), "Default Length should be 5")
	assert.Equal(uint16(0), pkt.PacketID(), "Default PacketID should be 0")
}

func TestUnsubackMarshal(t *testing.T) {
	pkt1 := NewUnsuback(RC_CONGESTION)
	pkt1.SetPacketID(1234)
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*Unsuback))
}

func TestUnsubackUnmarshalInvalid(t *testing.T) {
	assert := assert.New(t)

	// Packet too short.
	buff := bytes.NewBuffer([]byte{
		4,                   // Length
		byte(pkts.UNSUBACK), // Packet Type
		0, 1,                // Packet ID
		// Reason Code missing
	})
	_, err := ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad UNSUBACK2 packet length")
	}

	// Packet too long.
	buff = bytes.NewBuffer([]byte{
		5,                   // Length
		byte(pkts.UNSUBACK), // Packet Type
		0, 1,                // Packet ID
		0, 2, // Reason Code
		0, // junk
	})
	_, err = ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad UNSUBACK2 packet length")
	}
}

func TestUnsubackStringer(t *testing.T) {
	pkt := NewUnsuback(RC_CONGESTION)
	pkt.SetPacketID(1234)
	assert.Equal(t, "UNSUBACK2(ReasonCode=congestion, PacketID=1234)", pkt.String())
}
