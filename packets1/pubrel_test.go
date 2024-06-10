package packets1

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	pkts "github.com/energostack/bisquitt/packets"
)

func TestPubrelConstructor(t *testing.T) {
	assert := assert.New(t)

	pkt := NewPubrel()

	if pkt == nil {
		t.Fatal("New packet should not be nil")
	}

	assert.Equal("*packets1.Pubrel", reflect.TypeOf(pkt).String(), "Type should be Pubrel")
	assert.Equal(uint16(4), pkt.PacketLength(), "Default Length should be 4")
	assert.Equal(uint16(0), pkt.MessageID(), "Default MessageID should be 0")
}

func TestPubrelMarshal(t *testing.T) {
	pkt1 := NewPubrel()
	pkt1.SetMessageID(1234)
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*Pubrel))
}

func TestPubrelUnmarshalInvalid(t *testing.T) {
	assert := assert.New(t)

	// Packet too short.
	buff := bytes.NewBuffer([]byte{
		3,                 // Length
		byte(pkts.PUBREL), // MsgType
		0,                 // Message ID MSB
		// Message ID LSB missing
	})
	_, err := ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad PUBREL packet length")
	}

	// Packet too long.
	buff = bytes.NewBuffer([]byte{
		5,                 // Length
		byte(pkts.PUBREL), // MsgType
		0, 1,              // Message ID
		0, // junk
	})
	_, err = ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad PUBREL packet length")
	}
}

func TestPubrelStringer(t *testing.T) {
	pkt := NewPubrel()
	pkt.SetMessageID(1234)
	assert.Equal(t, "PUBREL(MessageID=1234)", pkt.String())
}
