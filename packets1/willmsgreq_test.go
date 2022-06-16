package packets1

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWillMsgReqStruct(t *testing.T) {
	msg := NewWillMsgReqMessage()

	if assert.NotNil(t, msg, "New message should not be nil") {
		assert.Equal(t, "*packets1.WillMsgReqMessage", reflect.TypeOf(msg).String(), "Type should be WillMsgReqMessage")
	}
}

func TestWillMsgReqMarshal(t *testing.T) {
	assert := assert.New(t)
	buf := bytes.NewBuffer(nil)

	msg1 := NewWillMsgReqMessage()
	if err := msg1.Write(buf); err != nil {
		t.Fatal(err)
	}

	r := bytes.NewReader(buf.Bytes())
	msg2, err := ReadPacket(r)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(msg1, msg2.(*WillMsgReqMessage))
}
