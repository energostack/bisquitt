package packets2

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	pkts "github.com/energomonitor/bisquitt/packets"
)

func TestDisconnectConstructor(t *testing.T) {
	assert := assert.New(t)

	reasonCode := RC_CONGESTION
	sessExpiry := uint32(123456)
	reasonString := "test reason"
	pkt := NewDisconnect(reasonCode, sessExpiry, reasonString)

	if pkt == nil {
		t.Fatal("New packet should not be nil")
	}

	assert.Equal("*packets2.Disconnect", reflect.TypeOf(pkt).String(), "Type should be Disconnect")
	assert.Equal(reasonCode, pkt.ReasonCode, "Bad ReasonCode value")
	assert.Equal(sessExpiry, pkt.SessionExpiryInterval, "Bad SessionExpiryInterval value")
	assert.Equal(reasonString, pkt.ReasonString, "Bad ReasonString value")
}

func TestDisconnectMarshal(t *testing.T) {
	pkt1 := NewDisconnect(RC_CONGESTION, uint32(123456), "test reason")
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*Disconnect))
}

func TestDisconnectUnmarshal(t *testing.T) {
	// Packet too short.
	buff := bytes.NewBuffer([]byte{
		6,                     // Length
		byte(pkts.DISCONNECT), // Packet Type
		0,                     // Reason Code
		0, 0, 0,               // Session Expiry Interval (without LSB)
		// Session Expiry Interval LSB missing
	})
	_, err := ReadPacket(buff)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "bad DISCONNECT2 packet length")
	}
}

func TestDisconnectStringer(t *testing.T) {
	pkt := NewDisconnect(RC_CONGESTION, uint32(123456), "test reason")
	assert.Equal(t, `DISCONNECT2(ReasonCode=congestion, Reason="test reason", SessExpiry=123456)`, pkt.String())
}
