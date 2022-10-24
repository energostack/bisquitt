package packets2

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	pkts "github.com/energomonitor/bisquitt/packets"
)

func TestPubrelConstructor(t *testing.T) {
	assert := assert.New(t)

	pkt := NewPubrel()

	if pkt == nil {
		t.Fatal("New packet should not be nil")
	}

	assert.Equal("*packets2.Pubrel", reflect.TypeOf(pkt).String(), "Type should be Pubrel")
	assert.Equal(uint16(4), pkt.PacketLength(), "Default Length should be 4")
	assert.Equal(uint16(0), pkt.PacketID(), "Default PacketID should be 0")
}

func TestPubrelMarshal(t *testing.T) {
	pkt1 := NewPubrel()
	pkt1.SetPacketID(1234)
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*Pubrel))
}

func TestPubrelUnmarshalInvalid(t *testing.T) {
	assert := assert.New(t)

	// Packet too short.
	buff := bytes.NewBuffer([]byte{
		3,                 // Length
		byte(pkts.PUBREL), // PktType
		0,                 // Packet ID MSB
		// Packet ID LSB missing
	})
	_, err := ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad PUBREL2 packet length")
	}

	// Packet too long.
	buff = bytes.NewBuffer([]byte{
		5,                 // Length
		byte(pkts.PUBREL), // PktType
		0, 1,              // Packet ID
		0, // junk
	})
	_, err = ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad PUBREL2 packet length")
	}
}

func TestPubrelStringer(t *testing.T) {
	pkt := NewPubrel()
	pkt.SetPacketID(1234)
	assert.Equal(t, "PUBREL2(PacketID=1234)", pkt.String())
}
