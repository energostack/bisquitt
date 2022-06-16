package packets1

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSubackStruct(t *testing.T) {
	topicID := uint16(12)
	qos := uint8(1)
	returnCode := RC_ACCEPTED
	msg := NewSuback(topicID, qos, returnCode)

	if assert.NotNil(t, msg, "New message should not be nil") {
		assert.Equal(t, "*packets1.Suback", reflect.TypeOf(msg).String(), "Type should be Suback")
		assert.Equal(t, uint16(8), msg.MessageLength(), "Default Length should be 8")
		assert.Equal(t, qos, msg.QOS, "Bad QOS value")
		assert.Equal(t, RC_ACCEPTED, msg.ReturnCode, "ReturnCode should be RC_ACCEPTED")
		assert.Equal(t, topicID, msg.TopicID, "Bad TopicID value")
		assert.Equal(t, uint16(0), msg.MessageID(), "Default MessageID should be 0")
	}
}

func TestSubackMarshal(t *testing.T) {
	assert := assert.New(t)
	buf := bytes.NewBuffer(nil)

	msg1 := NewSuback(123, 1, RC_CONGESTION)
	msg1.SetMessageID(12)
	if err := msg1.Write(buf); err != nil {
		t.Fatal(err)
	}

	r := bytes.NewReader(buf.Bytes())
	msg2, err := ReadPacket(r)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(msg1, msg2.(*Suback))
}
