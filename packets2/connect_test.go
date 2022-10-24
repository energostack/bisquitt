package packets2

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	pkts "github.com/energomonitor/bisquitt/packets"
)

func TestConnectConstructor(t *testing.T) {
	assert := assert.New(t)

	keepAlive := uint16(1234)
	sessExpiry := uint32(123456)
	maxPktSize := uint16(2345)
	clientID := "test-client"
	auth := true
	will := true
	cleanStart := true
	pkt := NewConnect(keepAlive, sessExpiry, maxPktSize, clientID, auth, will, cleanStart)

	if pkt == nil {
		t.Fatal("New packet should not be nil")
	}

	assert.Equal("*packets2.Connect", reflect.TypeOf(pkt).String(), "Type should be Connect")
	assert.Equal(keepAlive, pkt.KeepAlive, "Bad KeepAlive value")
	assert.Equal(sessExpiry, pkt.SessionExpiryInterval, "Bad SessionExpiryInterval value")
	assert.Equal(maxPktSize, pkt.MaxPacketSize, "Bad MaxPacketSize value")
	assert.Equal(clientID, pkt.ClientIdentifier, "Bad ClientIdentifier value")
	assert.Equal(auth, pkt.Authentication, "Bad Authentication value")
	assert.Equal(will, pkt.Will, "Bad Will value")
	assert.Equal(cleanStart, pkt.CleanStart, "Bad CleanStart value")
	assert.Equal(uint8(2), pkt.ProtocolVersion, "ProtocolVersion should be 2")
}

func TestConnectMarshal(t *testing.T) {
	pkt1 := NewConnect(1234, 123456, 2345, "test-client", true, true, true)
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*Connect))
}

func TestConnectUnmarshalInvalid(t *testing.T) {
	// Packet too short.
	buff := bytes.NewBuffer([]byte{
		12,                 // Length
		byte(pkts.CONNECT), // Packet Type
		0,                  // Flags
		2,                  // Protocol Version
		0, 0,               // Keep Alive
		0, 0, 0, 0, // Session Expiry Interval
		0, 0, // Max Packet Size
		// Client Identifier missing
	})
	_, err := ReadPacket(buff)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "bad CONNECT2 packet length")
	}

	// Invalid protocol version.
	buff = bytes.NewBuffer([]byte{
		13,                 // Length
		byte(pkts.CONNECT), // Packet Type
		0,                  // Flags
		1,                  // Protocol Version
		0, 0,               // Keep Alive
		0, 0, 0, 0, // Session Expiry Interval
		0, 0, // Max Packet Size
		byte('a'), // Client Identifier
	})
	_, err = ReadPacket(buff)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "bad CONNECT2 ProtocolVersion")
	}
}

func TestConnectStringer(t *testing.T) {
	pkt := NewConnect(1234, 123456, 2345, "test-client", true, true, true)
	assert.Equal(t, `CONNECT2(ClientID="test-client", CleanStart=true, Will=true, Auth=true, SessExpiry=123456, KeepAlive=1234, MaxPktSize=2345)`, pkt.String())
}
