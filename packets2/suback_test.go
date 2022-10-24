package packets2

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	pkts "github.com/energomonitor/bisquitt/packets"
)

func TestSubackConstructor(t *testing.T) {
	assert := assert.New(t)

	topicAlias := uint16(12)
	reasonCode := RC_CONGESTION
	grantedQOS := uint8(1)
	aliasType := TAT_PREDEFINED
	pkt := NewSuback(topicAlias, reasonCode, grantedQOS, aliasType)

	if pkt == nil {
		t.Fatal("New packet should not be nil")
	}

	assert.Equal("*packets2.Suback", reflect.TypeOf(pkt).String(), "Type should be Suback")
	assert.Equal(reasonCode, pkt.ReasonCode, "Bad ReasonCode value")
	assert.Equal(grantedQOS, pkt.GrantedQOS, "Bad GrantedQOS value")
	assert.Equal(topicAlias, pkt.TopicAlias, "Bad TopicAlias value")
	assert.Equal(aliasType, pkt.TopicAliasType, "Bad TopicAliasType value")
	assert.Equal(uint16(0), pkt.PacketID(), "Default PacketID should be 0")
	assert.Equal(uint16(8), pkt.PacketLength(), "Default Length should be 8")
}

func TestSubackMarshal(t *testing.T) {
	pkt1 := NewSuback(1234, RC_CONGESTION, 1, TAT_PREDEFINED)
	pkt1.SetPacketID(2345)
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*Suback))
}

func TestSubackUnmarshalInvalid(t *testing.T) {
	assert := assert.New(t)

	// Packet too short.
	buff := bytes.NewBuffer([]byte{
		7,                 // Length
		byte(pkts.SUBACK), // Packet Type
		0,                 // Flags
		0, 1,              // Topic Alias
		0, 2, // Packet ID
		// Reason Code missing
	})
	_, err := ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad SUBACK2 packet length")
	}

	// Packet too long.
	buff = bytes.NewBuffer([]byte{
		9,                 // Length
		byte(pkts.SUBACK), // Packet Type
		0,                 // Flags
		0, 1,              // Topic Alias
		0, 2, // Packet ID
		0, // Reason Code
		0, // junk
	})
	_, err = ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad SUBACK2 packet length")
	}
}

func TestSubackStringer(t *testing.T) {
	pkt := NewSuback(1234, RC_CONGESTION, 1, TAT_PREDEFINED)
	pkt.SetPacketID(2345)
	assert.Equal(t, "SUBACK2(Alias(predefined)=1234, ReasonCode=1, QOS=1, PacketID=2345)", pkt.String())
}
