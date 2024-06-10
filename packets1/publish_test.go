package packets1

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	pkts "github.com/energostack/bisquitt/packets"
)

func TestPublishConstructor(t *testing.T) {
	assert := assert.New(t)

	dup := true
	qos := uint8(1)
	retain := true
	topicIDType := TIT_SHORT
	topicID := uint16(1234)
	data := []byte("test-data")
	pkt := NewPublish(topicID, data, dup, qos, retain, topicIDType)

	if pkt == nil {
		t.Fatal("New packet should not be nil")
	}

	assert.Equal("*packets1.Publish", reflect.TypeOf(pkt).String(), "Type should be Publish")
	assert.Equal(dup, pkt.DUP(), "Bad Dup flag value")
	assert.Equal(retain, pkt.Retain, "Bad Retain flag value")
	assert.Equal(qos, pkt.QOS, "Bad QOS value")
	assert.Equal(topicIDType, pkt.TopicIDType, "Bad TopicIDType value")
	assert.Equal(topicID, pkt.TopicID, "Bad TopicID value")
	assert.Equal(uint16(0), pkt.MessageID(), "Default MessageID should be 0")
	assert.Equal(data, pkt.Data, "Bad Data value")
}

func TestPublishMarshal(t *testing.T) {
	pkt1 := NewPublish(1234, []byte("test-data"), true, 1, true, TIT_PREDEFINED)
	pkt1.SetMessageID(2345)
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*Publish))
}

func TestPublishUnmarshalInvalid(t *testing.T) {
	assert := assert.New(t)

	// Packet too short.
	buff := bytes.NewBuffer([]byte{
		3,                  // Length
		byte(pkts.PUBLISH), // MsgType
		0,                  // Flags
		0, 1,               // Topic ID
		0, // Message ID MSB
		// Message ID LSB missing
	})
	_, err := ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad PUBLISH packet length")
	}
}

func TestPublishStringer(t *testing.T) {
	pkt := NewPublish(1234, []byte("test-data"), true, 1, true, TIT_REGISTERED)
	pkt.SetMessageID(2345)
	assert.Equal(t,
		`PUBLISH(TopicID(r)=1234, Data="test-data", QOS=1, Retain=true, MessageID=2345, Dup=true)`,
		pkt.String())

	pkt = NewPublish(1234, []byte("test-data"), true, 1, true, TIT_PREDEFINED)
	pkt.SetMessageID(2345)
	assert.Equal(t,
		`PUBLISH(TopicID(p)=1234, Data="test-data", QOS=1, Retain=true, MessageID=2345, Dup=true)`,
		pkt.String())

	pkt = NewPublish(1234, []byte("test-data"), true, 1, true, TIT_SHORT)
	pkt.SetMessageID(2345)
	assert.Equal(t,
		`PUBLISH(TopicID(s)=1234, Data="test-data", QOS=1, Retain=true, MessageID=2345, Dup=true)`,
		pkt.String())
}
