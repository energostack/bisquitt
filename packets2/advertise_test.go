package packets2

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	pkts "github.com/energomonitor/bisquitt/packets"
)

func TestAdvertiseConstructor(t *testing.T) {
	assert := assert.New(t)

	gatewayID := uint8(12)
	duration := uint16(123)
	pkt := NewAdvertise(gatewayID, duration)

	if pkt == nil {
		t.Fatal("New packet should not be nil")
	}

	assert.Equal("*packets2.Advertise", reflect.TypeOf(pkt).String(), "Type should be Advertise")
	assert.Equal(gatewayID, pkt.GatewayID, "Bad GatewayID")
	assert.Equal(duration, pkt.Duration, "Bad Duration value")
	assert.Equal(uint16(5), pkt.PacketLength(), "Length should be 5")
}

func TestAdvertiseMarshal(t *testing.T) {
	pkt1 := NewAdvertise(12, 123)
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*Advertise))
}

func TestAdvertiseUnmarshalInvalid(t *testing.T) {
	assert := assert.New(t)

	// Packet too short.
	buff := bytes.NewBuffer([]byte{
		4,                    // Length
		byte(pkts.ADVERTISE), // MsgType
		0,                    // GwId
		0,                    // Duration - MSB
		// Duration - LSB missing
	})
	_, err := ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad ADVERTISE2 packet length")
	}

	// Packet too long.
	buff = bytes.NewBuffer([]byte{
		6,                    // Length
		byte(pkts.ADVERTISE), // MsgType
		0,                    // GwId
		0, 0,                 // Duration
		0, // junk byte
	})
	_, err = ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad ADVERTISE2 packet length")
	}
}

func TestAdvertiseStringer(t *testing.T) {
	pkt := NewAdvertise(12, 123)
	assert.Equal(t, "ADVERTISE2(GatewayID=12,Duration=123)", pkt.String())
}
