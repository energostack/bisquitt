package packets1

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWillTopicRespStruct(t *testing.T) {
	msg := NewWillTopicResp(RC_ACCEPTED)

	if assert.NotNil(t, msg, "New message should not be nil") {
		assert.Equal(t, "*packets1.WillTopicResp", reflect.TypeOf(msg).String(), "Type should be WillTopicResp")
		assert.Equal(t, RC_ACCEPTED, msg.ReturnCode, "ReturnCode should be RC_ACCEPTED")
	}
}

func TestWillTopicRespMarshal(t *testing.T) {
	assert := assert.New(t)
	buf := bytes.NewBuffer(nil)

	msg1 := NewWillTopicResp(RC_CONGESTION)
	if err := msg1.Write(buf); err != nil {
		t.Fatal(err)
	}

	r := bytes.NewReader(buf.Bytes())
	msg2, err := ReadPacket(r)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(msg1, msg2.(*WillTopicResp))
}
