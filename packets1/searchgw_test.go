package packets1

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	pkts "github.com/energomonitor/bisquitt/packets"
)

func TestSearchGwConstructor(t *testing.T) {
	assert := assert.New(t)

	radius := uint8(123)
	pkt := NewSearchGw(radius)

	if pkt == nil {
		t.Fatal("New packet should not be nil")
	}

	assert.Equal("*packets1.SearchGw", reflect.TypeOf(pkt).String(), "Type should be SearchGw")
	assert.Equal(uint16(3), pkt.PacketLength(), "Default Length should be 3")
	assert.Equal(radius, pkt.Radius, "Bad Radius value")
}

func TestSearchGwMarshal(t *testing.T) {
	pkt1 := NewSearchGw(123)
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*SearchGw))
}

func TestSearchGwUnmarshalInvalid(t *testing.T) {
	assert := assert.New(t)

	// Packet too short.
	buff := bytes.NewBuffer([]byte{
		2,                   // Length
		byte(pkts.SEARCHGW), // MsgType
		// Radius missing
	})
	_, err := ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad SEARCHGW packet length")
	}

	// Packet too long.
	buff = bytes.NewBuffer([]byte{
		4,                   // Length
		byte(pkts.SEARCHGW), // MsgType
		1,                   // Radius
		0,                   // junk
	})
	_, err = ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad SEARCHGW packet length")
	}
}

func TestSearchGwStringer(t *testing.T) {
	pkt := NewSearchGw(123)
	assert.Equal(t, "SEARCHGW(Radius=123)", pkt.String())
}
