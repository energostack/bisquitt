package packets1

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWillTopicReqStruct(t *testing.T) {
	pkt := NewWillTopicReq()

	if assert.NotNil(t, pkt, "New packet should not be nil") {
		assert.Equal(t, "*packets1.WillTopicReq", reflect.TypeOf(pkt).String(), "Type should be WillTopicReq")
	}
}

func TestWillTopicReqMarshal(t *testing.T) {
	assert := assert.New(t)
	buf := bytes.NewBuffer(nil)

	pkt1 := NewWillTopicReq()
	if err := pkt1.Write(buf); err != nil {
		t.Fatal(err)
	}

	r := bytes.NewReader(buf.Bytes())
	pkt2, err := ReadPacket(r)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(pkt1, pkt2.(*WillTopicReq))
}
