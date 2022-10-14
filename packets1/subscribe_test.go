package packets1

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	pkts "github.com/energomonitor/bisquitt/packets"
)

func TestSubscribeConstructor(t *testing.T) {
	assert := assert.New(t)

	topicID := uint16(12)
	topicIDType := TIT_REGISTERED
	topicName := "test-topic"
	qos := uint8(1)
	dup := true
	pkt := NewSubscribe(topicName, topicID, dup, qos, topicIDType)

	if pkt == nil {
		t.Fatal("New packet should not be nil")
	}

	assert.Equal("*packets1.Subscribe", reflect.TypeOf(pkt).String(), "Type should be Subscribe")
	assert.Equal(dup, pkt.DUP(), "Bad Dup flag value")
	assert.Equal(qos, pkt.QOS, "Bad QOS value")
	assert.Equal(topicIDType, pkt.TopicIDType, "Bad TopicIDType value")
	assert.Equal(topicID, pkt.TopicID, "Bad TopicID value")
	assert.Equal(uint16(0), pkt.MessageID(), "Default MessageID should be 0")
	assert.Equal(topicName, pkt.TopicName, "Bad Topicname value")
}

func TestSubscribeMarshal(t *testing.T) {
	// String topic ID.
	pkt1 := NewSubscribe("test-topic", 0, true, 1, TIT_STRING)
	pkt1.SetMessageID(1234)
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*Subscribe))

	// Predefined topic ID.
	pkt1 = NewSubscribe("", 1234, true, 1, TIT_PREDEFINED)
	pkt1.SetMessageID(2345)
	pkt2 = testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*Subscribe))

	// Short topic ID.
	pkt1 = NewSubscribe("", 1234, true, 1, TIT_SHORT)
	pkt1.SetMessageID(2345)
	pkt2 = testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*Subscribe))
}

func TestSubscribeUnmarshalInvalid(t *testing.T) {
	assert := assert.New(t)

	// Packet too short - string Topic ID.
	buff := bytes.NewBuffer([]byte{
		5,                    // Length
		byte(pkts.SUBSCRIBE), // MsgType
		TIT_STRING,           // Flags
		0, 1,                 // Message ID
		// Missing Topic Name
	})
	_, err := ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad SUBSCRIBE packet length")
	}

	// Packet too short - predefined Topic ID.
	buff = bytes.NewBuffer([]byte{
		5,                    // Length
		byte(pkts.SUBSCRIBE), // MsgType
		TIT_PREDEFINED,       // Flags
		0, 1,                 // Message ID
		0, // Topic ID MSB
		// Topic ID LSB missing
	})
	_, err = ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad SUBSCRIBE packet length")
	}

	// Packet too short - short Topic ID.
	buff = bytes.NewBuffer([]byte{
		5,                    // Length
		byte(pkts.SUBSCRIBE), // MsgType
		TIT_SHORT,            // Flags
		0, 1,                 // Message ID
		0, // Topic ID MSB
		// Topic ID LSB missing
	})
	_, err = ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad SUBSCRIBE packet length")
	}

	// Packet too long - predefined Topic ID.
	buff = bytes.NewBuffer([]byte{
		8,                    // Length
		byte(pkts.SUBSCRIBE), // MsgType
		TIT_PREDEFINED,       // Flags
		0, 1,                 // Message ID
		0, 2, // Topic ID
		0, // junk
	})
	_, err = ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad SUBSCRIBE packet length")
	}

	// Packet too long - short Topic ID.
	buff = bytes.NewBuffer([]byte{
		8,                    // Length
		byte(pkts.SUBSCRIBE), // MsgType
		TIT_SHORT,            // Flags
		0, 1,                 // Message ID
		0, 2, // Topic ID
		0, // junk
	})
	_, err = ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad SUBSCRIBE packet length")
	}

	// Invalid Topic ID Type.
	buff = bytes.NewBuffer([]byte{
		7,                    // Length
		byte(pkts.SUBSCRIBE), // MsgType
		3,                    // Flags (invalid Topic ID Type)
		0, 1,                 // Message ID
		0, 2, // Topic ID
	})
	_, err = ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "invalid TopicIDType")
	}
}

func TestSubscribeStringer(t *testing.T) {
	// String topic ID.
	pkt := NewSubscribe("test-topic", 0, true, 1, TIT_STRING)
	pkt.SetMessageID(1234)
	assert.Equal(t, `SUBSCRIBE(TopicName="test-topic", QOS=1, MessageID=1234, Dup=true)`, pkt.String())

	// Predefined topic ID.
	pkt = NewSubscribe("", 1234, true, 1, TIT_PREDEFINED)
	pkt.SetMessageID(2345)
	assert.Equal(t, "SUBSCRIBE(TopicID(p)=1234, QOS=1, MessageID=2345, Dup=true)", pkt.String())

	// Short topic ID.
	pkt = NewSubscribe("", pkts.EncodeShortTopic("ab"), true, 1, TIT_SHORT)
	pkt.SetMessageID(2345)
	assert.Equal(t, `SUBSCRIBE(TopicName(s)="ab", QOS=1, MessageID=2345, Dup=true)`, pkt.String())

	// Invalid Topic ID Type
	pkt = NewSubscribe("test-topic", 1234, true, 1, 3)
	pkt.SetMessageID(2345)
	assert.Equal(t, `SUBSCRIBE(TopicName="test-topic", QOS=1, TopicID=1234, TopicIDType=3 (INVALID!), MessageID=2345, Dup=true)`, pkt.String())
}
