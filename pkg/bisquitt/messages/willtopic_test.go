package messages

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWillTopicStruct(t *testing.T) {
	qos := uint8(1)
	retain := false
	willTopic := "test-topic"
	msg := NewWillTopicMessage(willTopic, qos, retain)

	if assert.NotNil(t, msg, "New message should not be nil") {
		assert.Equal(t, "*messages.WillTopicMessage", reflect.TypeOf(msg).String(), "Type should be WillTopicMessage")
		assert.Equal(t, qos, msg.QOS, "Bad QOS value")
		assert.Equal(t, retain, msg.Retain, "Bad Retain flag value")
		assert.Equal(t, willTopic, msg.WillTopic, "Bad WillTopic value")
	}
}

func TestWillTopicMarshal(t *testing.T) {
	assert := assert.New(t)
	buf := bytes.NewBuffer(nil)

	msg1 := NewWillTopicMessage("test-topic", 1, true)
	if err := msg1.Write(buf); err != nil {
		t.Fatal(err)
	}

	r := bytes.NewReader(buf.Bytes())
	msg2, err := ReadPacket(r)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(msg1, msg2.(*WillTopicMessage))
}
