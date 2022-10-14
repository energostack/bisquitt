package packets1

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	pkts "github.com/energomonitor/bisquitt/packets"
)

func TestConnackConstructor(t *testing.T) {
	assert := assert.New(t)

	returnCode := RC_CONGESTION
	pkt := NewConnack(returnCode)

	if pkt == nil {
		t.Fatal("New packet should not be nil")
	}

	assert.Equal("*packets1.Connack", reflect.TypeOf(pkt).String(), "Type should be Connack")
	assert.Equal(returnCode, pkt.ReturnCode, "invalid ReturnCode")
	assert.Equal(uint16(3), pkt.PacketLength(), "Length should be 3")
}

func TestConnackMarshal(t *testing.T) {
	pkt1 := NewConnack(RC_CONGESTION)
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*Connack))
}

func TestConnackUnmarshalInvalid(t *testing.T) {
	assert := assert.New(t)

	// Packet too short.
	buff := bytes.NewBuffer([]byte{
		2,                  // Length
		byte(pkts.CONNACK), // MsgType
		// Return Code missing
	})
	_, err := ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad CONNACK packet length")
	}

	// Packet too long.
	buff = bytes.NewBuffer([]byte{
		4,                  // Length
		byte(pkts.CONNACK), // MsgType
		0,                  // Return Code
		0,                  // junk
	})
	_, err = ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad CONNACK packet length")
	}
}

func TestConnackStringer(t *testing.T) {
	pkt := NewConnack(RC_CONGESTION)
	assert.Equal(t, "CONNACK(ReturnCode=1)", pkt.String())
}
