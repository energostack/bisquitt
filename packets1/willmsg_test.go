package packets1

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWillMsgStruct(t *testing.T) {
	payload := []byte("test-payload")
	msg := NewWillMsg(payload)

	if assert.NotNil(t, msg, "New packet should not be nil") {
		assert.Equal(t, "*packets1.WillMsg", reflect.TypeOf(msg).String(), "Type should be WillMsg")
		assert.Equal(t, payload, msg.WillMsg, "Bad WillMsg value")
	}
}

func TestWillMsgMarshal(t *testing.T) {
	assert := assert.New(t)
	buf := bytes.NewBuffer(nil)

	msg1 := NewWillMsg([]byte("test-message"))
	if err := msg1.Write(buf); err != nil {
		t.Fatal(err)
	}

	r := bytes.NewReader(buf.Bytes())
	msg2, err := ReadPacket(r)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(msg1, msg2.(*WillMsg))
}
