package packets2

import (
	"bytes"
	"reflect"
	"testing"

	pkts "github.com/energomonitor/bisquitt/packets"
	"github.com/stretchr/testify/assert"
)

func TestPingreqConstructor(t *testing.T) {
	assert := assert.New(t)

	maxMessages := uint8(123)
	clientID := "test-client"
	pkt := NewPingreq(maxMessages, clientID)

	if pkt == nil {
		t.Fatal("New packet should not be nil")
	}

	assert.Equal("*packets2.Pingreq", reflect.TypeOf(pkt).String(), "Type should be Pingreq")
	assert.Equal(maxMessages, pkt.MaxMessages, "Bad MaxMessages value")
	assert.Equal(clientID, pkt.ClientID, "Bad ClientID value")
}

func TestPingreqMarshal(t *testing.T) {
	// Packet without MaxMessages and ClientID.
	pkt1 := NewPingreq(0, "")
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*Pingreq))

	// Packet with MaxMessages and ClientID.
	pkt1 = NewPingreq(123, "test-client")
	pkt2 = testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*Pingreq))
}

func TestPingreqUnmarshal(t *testing.T) {
	assert := assert.New(t)

	// Packet without MaxMessages and ClientID is correct (length=2).
	buff := bytes.NewBuffer([]byte{
		2,                  // Length
		byte(pkts.PINGREQ), // Packet Type
	})
	_, err := ReadPacket(buff)
	assert.Nil(err)

	// Packet too short (length=3).
	buff = bytes.NewBuffer([]byte{
		3,                  // Length
		byte(pkts.PINGREQ), // Packet Type
		0,                  // Max Messages
		// Client Identifier missing
	})
	_, err = ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad PINGREQ2 packet length")
	}

	// Packet with MaxMessages and ClientID is correct (length=4).
	buff = bytes.NewBuffer([]byte{
		4,                  // Length
		byte(pkts.PINGREQ), // Packet Type
		0,                  // Max Messages
		byte('a'),          // Client Identifier
	})
	_, err = ReadPacket(buff)
	assert.Nil(err)
}

func TestPingreqStringer(t *testing.T) {
	// Packet without MaxMessages and ClientID.
	pkt := NewPingreq(0, "")
	assert.Equal(t, `PINGREQ2`, pkt.String())

	// Packet with MaxMessages and ClientID.
	pkt = NewPingreq(123, "test-client")
	assert.Equal(t, `PINGREQ2(MaxMessages=123, ClientID="test-client")`, pkt.String())
}
