package packets1

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPingreqStruct(t *testing.T) {
	clientID := []byte("test-client")
	msg := NewPingreqMessage(clientID)

	if assert.NotNil(t, msg, "New message should not be nil") {
		assert.Equal(t, "*packets1.PingreqMessage", reflect.TypeOf(msg).String(), "Type should be PingreqMessage")
		assert.Equal(t, clientID, msg.ClientID, "Bad ClientID value")
	}
}

func TestPingreqMarshal(t *testing.T) {
	assert := assert.New(t)
	buf := bytes.NewBuffer(nil)

	msg1 := NewPingreqMessage([]byte("test-client"))
	if err := msg1.Write(buf); err != nil {
		t.Fatal(err)
	}

	r := bytes.NewReader(buf.Bytes())
	msg2, err := ReadPacket(r)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(msg1, msg2.(*PingreqMessage))
}
