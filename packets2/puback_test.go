package packets2

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	pkts "github.com/energomonitor/bisquitt/packets"
)

func TestPubackConstructor(t *testing.T) {
	assert := assert.New(t)

	pkt := NewPuback(RC_CONGESTION)

	if pkt == nil {
		t.Fatal("New packet should not be nil")
	}

	assert.Equal("*packets2.Puback", reflect.TypeOf(pkt).String(), "Type should be Puback")
	assert.Equal(uint16(0), pkt.PacketID(), "Default PacketID should be 0")
	assert.Equal(RC_CONGESTION, pkt.ReasonCode, "ReasonCode should be RC_CONGESTION")
	assert.Equal(uint16(5), pkt.PacketLength(), "Default Length should be 5")
}

func TestPubackMarshal(t *testing.T) {
	pkt1 := NewPuback(RC_CONGESTION)
	pkt1.SetPacketID(123)
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*Puback))
}

func TestPubackUnmarshalInvalid(t *testing.T) {
	assert := assert.New(t)

	// Packet too short.
	buff := bytes.NewBuffer([]byte{
		4,                 // Length
		byte(pkts.PUBACK), // Packet Type
		0, 2,              // Packet ID
		// Reason Code missing
	})
	_, err := ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad PUBACK2 packet length")
	}

	// Packet too long.
	buff = bytes.NewBuffer([]byte{
		6,                 // Length
		byte(pkts.PUBACK), // Packet Type
		0, 2,              // Packet ID
		0, // Reason Code
		0, // junk
	})
	_, err = ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad PUBACK2 packet length")
	}
}

func TestPubackStringer(t *testing.T) {
	pkt := NewPuback(RC_CONGESTION)
	pkt.SetPacketID(123)
	assert.Equal(t, "PUBACK2(ReasonCode=congestion, PacketID=123)", pkt.String())
}
