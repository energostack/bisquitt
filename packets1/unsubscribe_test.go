package packets1

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	pkts "github.com/energomonitor/bisquitt/packets"
)

func TestUnsubscribeConstructor(t *testing.T) {
	assert := assert.New(t)

	topicID := uint16(12)
	topicIDType := TIT_REGISTERED
	topicName := "test-topic"
	pkt := NewUnsubscribe(topicName, topicID, topicIDType)

	if pkt == nil {
		t.Fatal("New packet should not be nil")
	}

	assert.Equal("*packets1.Unsubscribe", reflect.TypeOf(pkt).String(), "Type should be Unsubscribe")
	assert.Equal(topicIDType, pkt.TopicIDType, "Bad TopicIDType value")
	assert.Equal(topicID, pkt.TopicID, "Bad TopicID value")
	assert.Equal(uint16(0), pkt.MessageID(), "Default MessageID should be 0")
	assert.Equal(topicName, pkt.TopicName, "Bad Topicname value")
}

func TestUnsubscribeMarshal(t *testing.T) {
	// String topic ID.
	pkt1 := NewUnsubscribe("test-topic", 0, TIT_STRING)
	pkt1.SetMessageID(1234)
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*Unsubscribe))

	// Predefined topic ID.
	pkt1 = NewUnsubscribe("", 1234, TIT_PREDEFINED)
	pkt1.SetMessageID(2345)
	pkt2 = testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*Unsubscribe))

	// Short topic ID.
	pkt1 = NewUnsubscribe("", 1234, TIT_SHORT)
	pkt1.SetMessageID(2345)
	pkt2 = testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*Unsubscribe))
}

func TestUnsubscribeUnmarshalInvalid(t *testing.T) {
	assert := assert.New(t)

	// Packet too short - string Topic ID.
	buff := bytes.NewBuffer([]byte{
		5,                      // Length
		byte(pkts.UNSUBSCRIBE), // MsgType
		TIT_STRING,             // Flags
		0, 1,                   // Message ID
		// Missing Topic Name
	})
	_, err := ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad UNSUBSCRIBE packet length")
	}

	// Packet too short - predefined Topic ID.
	buff = bytes.NewBuffer([]byte{
		5,                      // Length
		byte(pkts.UNSUBSCRIBE), // MsgType
		TIT_PREDEFINED,         // Flags
		0, 1,                   // Message ID
		0, // Topic ID MSB
		// Topic ID LSB missing
	})
	_, err = ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad UNSUBSCRIBE packet length")
	}

	// Packet too short - short Topic ID.
	buff = bytes.NewBuffer([]byte{
		5,                      // Length
		byte(pkts.UNSUBSCRIBE), // MsgType
		TIT_SHORT,              // Flags
		0, 1,                   // Message ID
		0, // Topic ID MSB
		// Topic ID LSB missing
	})
	_, err = ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad UNSUBSCRIBE packet length")
	}

	// Packet too long - predefined Topic ID.
	buff = bytes.NewBuffer([]byte{
		8,                      // Length
		byte(pkts.UNSUBSCRIBE), // MsgType
		TIT_PREDEFINED,         // Flags
		0, 1,                   // Message ID
		0, 2, // Topic ID
		0, // junk
	})
	_, err = ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad UNSUBSCRIBE packet length")
	}

	// Packet too long - short Topic ID.
	buff = bytes.NewBuffer([]byte{
		8,                      // Length
		byte(pkts.UNSUBSCRIBE), // MsgType
		TIT_SHORT,              // Flags
		0, 1,                   // Message ID
		0, 2, // Topic ID
		0, // junk
	})
	_, err = ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad UNSUBSCRIBE packet length")
	}

	// Invalid Topic ID Type.
	buff = bytes.NewBuffer([]byte{
		7,                      // Length
		byte(pkts.UNSUBSCRIBE), // MsgType
		3,                      // Flags (invalid Topic ID Type)
		0, 1,                   // Message ID
		0, 2, // Topic ID
	})
	_, err = ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "invalid TopicIDType")
	}
}

func TestUnsubscribeStringer(t *testing.T) {
	// String topic ID.
	pkt := NewUnsubscribe("test-topic", 0, TIT_STRING)
	pkt.SetMessageID(1234)
	assert.Equal(t, `UNSUBSCRIBE(TopicName="test-topic", MessageID=1234)`, pkt.String())

	// Predefined topic ID.
	pkt = NewUnsubscribe("", 1234, TIT_PREDEFINED)
	pkt.SetMessageID(2345)
	assert.Equal(t, "UNSUBSCRIBE(TopicID(p)=1234, MessageID=2345)", pkt.String())

	// Short topic ID.
	pkt = NewUnsubscribe("", pkts.EncodeShortTopic("ab"), TIT_SHORT)
	pkt.SetMessageID(2345)
	assert.Equal(t, `UNSUBSCRIBE(TopicName(s)="ab", MessageID=2345)`, pkt.String())

	// Invalid Topic ID Type
	pkt = NewUnsubscribe("test-topic", 1234, 3)
	pkt.SetMessageID(2345)
	assert.Equal(t, `UNSUBSCRIBE(TopicName="test-topic", TopicID=1234, TopicIDType=3 (INVALID!), MessageID=2345)`, pkt.String())
}
