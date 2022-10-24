package packets2

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	pkts "github.com/energomonitor/bisquitt/packets"
)

func TestRegisterConstructor(t *testing.T) {
	assert := assert.New(t)

	topicAlias := uint16(123)
	topic := "test-topic"
	pkt := NewRegister(topicAlias, topic)

	if pkt == nil {
		t.Fatal("New packet should not be nil")
	}

	assert.Equal("*packets2.Register", reflect.TypeOf(pkt).String(), "Type should be Register")
	assert.Equal(topicAlias, pkt.TopicAlias, "Bad TopicAlias value")
	assert.Equal(uint16(0), pkt.PacketID(), "Default PacketID should be 0")
	assert.Equal(topic, pkt.TopicName, "Bad TopicName value")
}

func TestRegisterMarshal(t *testing.T) {
	pkt1 := NewRegister(1234, "test-topic")
	pkt1.SetPacketID(2345)
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*Register))
}

func TestRegisterUnmarshalInvalid(t *testing.T) {
	assert := assert.New(t)

	// Packet too short.
	buff := bytes.NewBuffer([]byte{
		6,                   // Length
		byte(pkts.REGISTER), // Packet Type
		0, 1,                // Topic Alias
		0, 2, // Packet ID
		// TopicName missing
	})
	_, err := ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad REGISTER2 packet length")
	}
}

func TestRegisterStringer(t *testing.T) {
	pkt := NewRegister(1234, "test-topic")
	pkt.SetPacketID(2345)
	assert.Equal(t, `REGISTER2(Topic="test-topic", Alias=1234, PacketID=2345)`,
		pkt.String())
}
