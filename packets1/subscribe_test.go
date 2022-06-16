package packets1

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSubscribeStruct(t *testing.T) {
	topicID := uint16(12)
	topicIDType := TIT_REGISTERED
	topicName := []byte("test-topic")
	qos := uint8(1)
	dup := true
	msg := NewSubscribeMessage(topicID, topicIDType, topicName, qos, dup)

	if assert.NotNil(t, msg, "New message should not be nil") {
		assert.Equal(t, "*packets1.SubscribeMessage", reflect.TypeOf(msg).String(), "Type should be SubscribeMessage")
		assert.Equal(t, dup, msg.DUP(), "Bad Dup flag value")
		assert.Equal(t, qos, msg.QOS, "Bad QOS value")
		assert.Equal(t, topicIDType, msg.TopicIDType, "Bad TopicIDType value")
		assert.Equal(t, topicID, msg.TopicID, "Bad TopicID value")
		assert.Equal(t, uint16(0), msg.MessageID(), "Default MessageID should be 0")
		assert.Equal(t, topicName, msg.TopicName, "Bad Topicname value")
	}
}

func TestSubscribeMarshalString(t *testing.T) {
	assert := assert.New(t)
	buf := bytes.NewBuffer(nil)

	msg1 := NewSubscribeMessage(0, TIT_STRING, []byte("test-topic"), 1, true)
	msg1.SetMessageID(12)
	if err := msg1.Write(buf); err != nil {
		t.Fatal(err)
	}

	r := bytes.NewReader(buf.Bytes())
	msg2, err := ReadPacket(r)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(msg1, msg2.(*SubscribeMessage))
}

func TestSubscribeMarshalShort(t *testing.T) {
	assert := assert.New(t)
	buf := bytes.NewBuffer(nil)

	msg1 := NewSubscribeMessage(123, TIT_SHORT, nil, 1, true)
	msg1.SetMessageID(12)
	if err := msg1.Write(buf); err != nil {
		t.Fatal(err)
	}

	r := bytes.NewReader(buf.Bytes())
	msg2, err := ReadPacket(r)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(msg1, msg2.(*SubscribeMessage))
}
