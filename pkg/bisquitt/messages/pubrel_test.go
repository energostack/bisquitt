package messages

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPubrelStruct(t *testing.T) {
	msg := NewPubrelMessage()

	if assert.NotNil(t, msg, "New message should not be nil") {
		assert.Equal(t, "*messages.PubrelMessage", reflect.TypeOf(msg).String(), "Type should be PubrelMessage")
		assert.Equal(t, uint16(4), msg.MessageLength(), "Default Length should be 4")
		assert.Equal(t, uint16(0), msg.MessageID(), "Default MessageID should be 0")
	}
}

func TestPubrelMarshal(t *testing.T) {
	assert := assert.New(t)
	buf := bytes.NewBuffer(nil)

	msg1 := NewPubrelMessage()
	msg1.SetMessageID(12)
	if err := msg1.Write(buf); err != nil {
		t.Fatal(err)
	}

	r := bytes.NewReader(buf.Bytes())
	msg2, err := ReadPacket(r)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(msg1, msg2.(*PubrelMessage))
}
