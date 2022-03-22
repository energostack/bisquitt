package messages

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPingrespStruct(t *testing.T) {
	msg := NewPingrespMessage()

	if assert.NotNil(t, msg, "New message should not be nil") {
		assert.Equal(t, "*messages.PingrespMessage", reflect.TypeOf(msg).String(), "Type should be PingrespMessage")
		assert.Equal(t, uint16(2), msg.MessageLength(), "Default Length should be 2")
	}
}

func TestPingrespMarshal(t *testing.T) {
	assert := assert.New(t)
	buf := bytes.NewBuffer(nil)

	msg1 := NewPingrespMessage()
	if err := msg1.Write(buf); err != nil {
		t.Fatal(err)
	}

	r := bytes.NewReader(buf.Bytes())
	msg2, err := ReadPacket(r)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(msg1, msg2.(*PingrespMessage))
}
