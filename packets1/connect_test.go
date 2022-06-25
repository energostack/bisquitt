package packets1

import (
	"bytes"
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
	}
}

func TestConnectMarshal(t *testing.T) {
	assert := assert.New(t)
	buf := bytes.NewBuffer(nil)

	pkt1 := NewConnect([]byte("test-client"), true, true, 75)
	err := pkt1.Write(buf)
	if err != nil {
		t.Fatal(err)
	}

	r := bytes.NewReader(buf.Bytes())
	pkt2, err := ReadPacket(r)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(pkt1, pkt2.(*Connect))
}
