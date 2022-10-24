package packets2

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	pkts "github.com/energomonitor/bisquitt/packets"
)

func TestRegackConstructor(t *testing.T) {
	assert := assert.New(t)

	topicAlias := uint16(1234)
	reasonCode := RC_CONGESTION
	aliasType := TAT_PREDEFINED
	pkt := NewRegack(topicAlias, reasonCode, aliasType)

	if pkt == nil {
		t.Fatal("New packet should not be nil")
	}

	assert.Equal("*packets2.Regack", reflect.TypeOf(pkt).String(), "Type should be Regack")
	assert.Equal(topicAlias, pkt.TopicAlias, "Bad TopicAlias value")
	assert.Equal(reasonCode, pkt.ReasonCode, "Bad ReasonCode value")
	assert.Equal(aliasType, pkt.TopicAliasType, "Bad TopicAliasType value")
	assert.Equal(uint16(0), pkt.PacketID(), "Default PacketID should be 0")
	assert.Equal(uint16(8), pkt.PacketLength(), "Default Length should be 8")
}

func TestRegackMarshal(t *testing.T) {
	pkt1 := NewRegack(1234, RC_CONGESTION, TAT_PREDEFINED)
	pkt1.SetPacketID(2345)
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*Regack))
}

func TestRegackUnmarshalInvalid(t *testing.T) {
	assert := assert.New(t)

	// Packet too short.
	buff := bytes.NewBuffer([]byte{
		7,                 // Length
		byte(pkts.REGACK), // Packet Type
		0,                 // Flags
		0, 1,              // Topic Alias
		0, 2, // Packet ID
		// Reason Code missing
	})
	_, err := ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad REGACK2 packet length")
	}

	// Packet too long.
	buff = bytes.NewBuffer([]byte{
		7,                 // Length
		byte(pkts.REGACK), // Packet Type
		0,                 // Flags
		0, 1,              // Topic Alias
		0, 2, // Packet ID
		0, // Reason Code
		0, // junk
	})
	_, err = ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad REGACK2 packet length")
	}
}

func TestRegackStringer(t *testing.T) {
	pkt := NewRegack(1234, RC_CONGESTION, TAT_PREDEFINED)
	pkt.SetPacketID(2345)
	assert.Equal(t, "REGACK2(Alias(predefined)=1234, ReasonCode=1, PacketID=2345)", pkt.String())
}
