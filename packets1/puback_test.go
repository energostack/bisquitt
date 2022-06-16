package packets1

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPubackStruct(t *testing.T) {
	topicID := uint16(123)
	msg := NewPuback(topicID, RC_ACCEPTED)

	if assert.NotNil(t, msg, "New message should not be nil") {
		assert.Equal(t, "*packets1.Puback", reflect.TypeOf(msg).String(), "Type should be Puback")
		assert.Equal(t, topicID, msg.TopicID, fmt.Sprintf("TopicID should be %d", topicID))
		assert.Equal(t, uint16(0), msg.MessageID(), "Default MessageID should be 0")
		assert.Equal(t, RC_ACCEPTED, msg.ReturnCode, "ReturnCode should be RC_ACCEPTED")
		assert.Equal(t, uint16(7), msg.MessageLength(), "Default Length should be 2")
	}
}

func TestPubackMarshal(t *testing.T) {
	assert := assert.New(t)
	buf := bytes.NewBuffer(nil)

	msg1 := NewPuback(123, RC_CONGESTION)
	msg1.SetMessageID(12)
	if err := msg1.Write(buf); err != nil {
		t.Fatal(err)
	}

	r := bytes.NewReader(buf.Bytes())
	msg2, err := ReadPacket(r)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(msg1, msg2.(*Puback))
}
