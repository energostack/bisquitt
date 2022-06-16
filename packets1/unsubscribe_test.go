package packets1

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnsubscribeStruct(t *testing.T) {
	topicID := uint16(12)
	topicIDType := TIT_REGISTERED
	topicName := []byte("test-topic")
	msg := NewUnsubscribeMessage(topicID, topicIDType, topicName)

	if assert.NotNil(t, msg, "New message should not be nil") {
		assert.Equal(t, "*packets1.UnsubscribeMessage", reflect.TypeOf(msg).String(), "Type should be UnsubscribeMessage")
		assert.Equal(t, topicIDType, msg.TopicIDType, "Bad TopicIDType value")
		assert.Equal(t, topicID, msg.TopicID, "Bad TopicID value")
		assert.Equal(t, uint16(0), msg.MessageID(), "Default MessageID should be 0")
		assert.Equal(t, topicName, msg.TopicName, "Bad Topicname value")
	}
}

func TestUnsubscribeMarshalString(t *testing.T) {
	assert := assert.New(t)
	buf := bytes.NewBuffer(nil)

	msg1 := NewUnsubscribeMessage(0, TIT_STRING, []byte("test-topic"))
	msg1.SetMessageID(12)
	if err := msg1.Write(buf); err != nil {
		t.Fatal(err)
	}

	r := bytes.NewReader(buf.Bytes())
	msg2, err := ReadPacket(r)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(msg1, msg2.(*UnsubscribeMessage))
}

func TestUnsubscribeMarshalShort(t *testing.T) {
	assert := assert.New(t)
	buf := bytes.NewBuffer(nil)

	msg1 := NewUnsubscribeMessage(123, TIT_SHORT, nil)
	msg1.SetMessageID(12)
	if err := msg1.Write(buf); err != nil {
		t.Fatal(err)
	}

	r := bytes.NewReader(buf.Bytes())
	msg2, err := ReadPacket(r)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(msg1, msg2.(*UnsubscribeMessage))
}
