package messages

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConnackStruct(t *testing.T) {
	msg := NewConnackMessage(RC_ACCEPTED)

	if assert.NotNil(t, msg, "New message should not be nil") {
		assert.Equal(t, "*messages.ConnackMessage", reflect.TypeOf(msg).String(), "Type should be ConnackMessage")
		assert.Equal(t, RC_ACCEPTED, msg.ReturnCode, "ReturnCode should be RC_ACCEPTED")
		assert.Equal(t, uint16(3), msg.MessageLength(), "Length should be 3")
	}
}

func TestConnackMarshal(t *testing.T) {
	assert := assert.New(t)
	buf := bytes.NewBuffer(nil)

	msg1 := NewConnackMessage(RC_CONGESTION)
	if err := msg1.Write(buf); err != nil {
		t.Fatal(err)
	}

	r := bytes.NewReader(buf.Bytes())
	msg2, err := ReadPacket(r)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(msg1, msg2.(*ConnackMessage))
}
