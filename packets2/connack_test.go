package packets2

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	pkts "github.com/energomonitor/bisquitt/packets"
)

func TestConnackConstructor(t *testing.T) {
	assert := assert.New(t)

	reasonCode := RC_CONGESTION
	sessExpiry := uint32(1234)
	assignedClientID := "client-id"
	sessPresent := true
	pkt := NewConnack(reasonCode, sessExpiry, assignedClientID, sessPresent)

	if pkt == nil {
		t.Fatal("New packet should not be nil")
	}

	assert.Equal("*packets2.Connack", reflect.TypeOf(pkt).String(), "Type should be Connack")
	assert.Equal(reasonCode, pkt.ReasonCode, "Invalid ReasonCode")
	assert.Equal(sessExpiry, pkt.SessionExpiryInterval, "Invalid SessionExpiryInterval")
	assert.Equal(assignedClientID, pkt.AssignedClientIdentifier, "Invalid AssignedClientIdentifier")
	assert.Equal(2+connackHeaderLength+uint16(len(assignedClientID)), pkt.PacketLength(), "Invalid PacketLength")
}

func TestConnackMarshal(t *testing.T) {
	pkt1 := NewConnack(RC_CONGESTION, 1234, "client-id", true)
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*Connack))
}

func TestConnackUnmarshalInvalid(t *testing.T) {
	assert := assert.New(t)

	// Packet too short.
	buff := bytes.NewBuffer([]byte{
		7,                  // Length
		byte(pkts.CONNACK), // Packet Type
		0,                  // Reason Code
		0,                  // Flags
		0, 0, 0,            // Session Expiry Interval (without LSB)
		// Session Expiry Interval LSB missing
	})
	_, err := ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad CONNACK2 packet length")
	}
}

func TestConnackStringer(t *testing.T) {
	pkt := NewConnack(RC_CONGESTION, 1234, "client-id", true)
	assert.Equal(t, `CONNACK2(ReasonCode=1,AssClientId="client-id",SessExpiry=1234,SessPresent=true)`, pkt.String())
}
