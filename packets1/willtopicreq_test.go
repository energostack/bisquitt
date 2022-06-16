package packets1

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWillTopicReqStruct(t *testing.T) {
	msg := NewWillTopicReq()

	if assert.NotNil(t, msg, "New packet should not be nil") {
		assert.Equal(t, "*packets1.WillTopicReq", reflect.TypeOf(msg).String(), "Type should be WillTopicReq")
	}
}

func TestWillTopicReqMarshal(t *testing.T) {
	assert := assert.New(t)
	buf := bytes.NewBuffer(nil)

	msg1 := NewWillTopicReq()
	if err := msg1.Write(buf); err != nil {
		t.Fatal(err)
	}

	r := bytes.NewReader(buf.Bytes())
	msg2, err := ReadPacket(r)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(msg1, msg2.(*WillTopicReq))
}
