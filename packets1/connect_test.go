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
	pkt := NewConnect(clientID, cleanSession, will, duration)

	if assert.NotNil(t, pkt, "New packet should not be nil") {
		assert.Equal(t, "*packets1.Connect", reflect.TypeOf(pkt).String(), "Type should be Connect")
		assert.Equal(t, will, pkt.Will, "Bad Will value")
		assert.Equal(t, cleanSession, pkt.CleanSession, "Bad CleanSession value")
		assert.Equal(t, uint8(1), pkt.ProtocolID, "Default ProtocolID should be 1")
		assert.Equal(t, duration, pkt.Duration, "Bad Duration value")
		assert.Equal(t, clientID, pkt.ClientID, "Bad ClientID value")

		unpackError := pkt.Unpack([]byte(""))
		assert.NotNil(t, unpackError)
		assert.Error(t, unpackError)
		assert.Contains(t, unpackError.Error(), "bad CONNECT packet length")

		assert.Equal(t, "CONNECT(ClientID=\"client-id\", CleanSession=true, Will=true, Duration=90)", pkt.String())
	}
}

func TestConnectMarshal(t *testing.T) {
	pkt1 := NewConnect([]byte("test-client"), true, true, 75)
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*Connect))
}
