package packets1

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConnect(t *testing.T) {
	clientID := []byte("client-id")
	cleanSession := true
	will := true
	duration := uint16(90)
	pkt := NewConnect(duration, clientID, will, cleanSession)

	if assert.NotNil(t, pkt, "New packet should not be nil") {
		assert.Equal(t, "*packets1.Connect", reflect.TypeOf(pkt).String(), "Type should be Connect")
		assert.Equal(t, will, pkt.Will, "Bad Will value")
		assert.Equal(t, cleanSession, pkt.CleanSession, "Bad CleanSession value")
		assert.Equal(t, uint8(1), pkt.ProtocolID, "Default ProtocolID should be 1")
		assert.Equal(t, duration, pkt.Duration, "Bad Duration value")
		assert.Equal(t, clientID, pkt.ClientID, "Bad ClientID value")
	}
}

func TestConnectMarshal(t *testing.T) {
	pkt1 := NewConnect(75, []byte("test-client"), true, true)
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*Connect))
}
