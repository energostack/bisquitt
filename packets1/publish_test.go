package packets1

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPublishStruct(t *testing.T) {
	dup := true
	retain := true
	qos := uint8(1)
	topicIDType := TIT_SHORT
	topicID := uint16(123)
	payload := []byte("test-payload")
	msg := NewPublish(topicID, topicIDType, payload, qos, retain, dup)

	if assert.NotNil(t, msg, "New message should not be nil") {
		assert.Equal(t, "*packets1.Publish", reflect.TypeOf(msg).String(), "Type should be Publish")
		assert.Equal(t, dup, msg.DUP(), "Bad Dup flag value")
		assert.Equal(t, retain, msg.Retain, "Bad Retain flag value")
		assert.Equal(t, qos, msg.QOS, "Bad QOS value")
		assert.Equal(t, topicIDType, msg.TopicIDType, "Bad TopicIDType value")
		assert.Equal(t, topicID, msg.TopicID, "Bad TopicID value")
		assert.Equal(t, uint16(0), msg.MessageID(), "Default MessageID should be 0")
		assert.Equal(t, payload, msg.Data, "Bad Data value")
	}
}

func TestPublishMarshal(t *testing.T) {
	assert := assert.New(t)
	buf := bytes.NewBuffer(nil)

	msg1 := NewPublish(123, TIT_PREDEFINED,
		[]byte("test-payload"), 1, true, true)
	msg1.SetMessageID(12)
	if err := msg1.Write(buf); err != nil {
		t.Fatal(err)
	}

	r := bytes.NewReader(buf.Bytes())
	msg2, err := ReadPacket(r)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(msg1, msg2.(*Publish))
}
