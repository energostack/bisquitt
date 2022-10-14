package packets1

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	pkts "github.com/energomonitor/bisquitt/packets"
)

func TestRegisterConstructor(t *testing.T) {
	assert := assert.New(t)

	topicID := uint16(123)
	topic := "test-topic"
	pkt := NewRegister(topicID, topic)

	if pkt == nil {
		t.Fatal("New packet should not be nil")
	}

	assert.Equal("*packets1.Register", reflect.TypeOf(pkt).String(), "Type should be Register")
	assert.Equal(topicID, pkt.TopicID, "Bad TopicID value")
	assert.Equal(uint16(0), pkt.MessageID(), "Default MessageID should be 0")
	assert.Equal(topic, pkt.TopicName, "Bad TopicName value")
}

func TestRegisterMarshal(t *testing.T) {
	pkt1 := NewRegister(1234, "test-topic")
	pkt1.SetMessageID(2345)
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*Register))
}

func TestRegisterUnmarshalInvalid(t *testing.T) {
	assert := assert.New(t)

	// Packet too short.
	buff := bytes.NewBuffer([]byte{
		6,                   // Length
		byte(pkts.REGISTER), // MsgType
		0, 1,                // Topic ID
		0, 2, // Message ID
		// TopicName missing
	})
	_, err := ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad REGISTER packet length")
	}
}

func TestRegisterStringer(t *testing.T) {
	pkt := NewRegister(1234, "test-topic")
	pkt.SetMessageID(2345)
	assert.Equal(t, `REGISTER(TopicName="test-topic", TopicID=1234, MessageID=2345)`, pkt.String())
}
