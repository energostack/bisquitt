package packets1

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	pkts "github.com/energostack/bisquitt/packets"
)

func TestConnectConstructor(t *testing.T) {
	assert := assert.New(t)

	clientID := []byte("client-id")
	cleanSession := true
	will := true
	duration := uint16(90)
	pkt := NewConnect(duration, clientID, will, cleanSession)

	if pkt == nil {
		t.Fatal("New packet should not be nil")
	}

	assert.Equal("*packets1.Connect", reflect.TypeOf(pkt).String(), "Type should be Connect")
	assert.Equal(will, pkt.Will, "Bad Will value")
	assert.Equal(cleanSession, pkt.CleanSession, "Bad CleanSession value")
	assert.Equal(duration, pkt.Duration, "Bad Duration value")
	assert.Equal(clientID, pkt.ClientID, "Bad ClientID value")
}

func TestConnectMarshal(t *testing.T) {
	pkt1 := NewConnect(75, []byte("test-client"), true, true)
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*Connect))
}

func TestConnectUnmarshalInvalid(t *testing.T) {
	// Packet too short.
	buff := bytes.NewBuffer([]byte{
		6,                  // Length
		byte(pkts.CONNECT), // MsgType
		0,                  // Flags
		1,                  // ProtocolId
		0, 0,               // Duration
		// ClientId missing
	})
	_, err := ReadPacket(buff)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "bad CONNECT packet length")
	}

	// Invalid protocol version.
	buff = bytes.NewBuffer([]byte{
		6,                  // Length
		byte(pkts.CONNECT), // MsgType
		0,                  // Flags
		2,                  // ProtocolId
		0, 0,               // Duration
		byte('a'), // ClientId
	})
	_, err = ReadPacket(buff)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "bad CONNECT ProtocolId")
	}
}

func TestConnectStringer(t *testing.T) {
	pkt := NewConnect(123, []byte("client-id"), true, true)
	assert.Equal(t, "CONNECT(ClientID=\"client-id\", CleanSession=true, Will=true, Duration=123)", pkt.String())
}
