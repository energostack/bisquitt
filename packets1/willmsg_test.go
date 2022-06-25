package packets1

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWillMsgStruct(t *testing.T) {
	payload := []byte("test-payload")
	pkt := NewWillMsg(payload)

	if assert.NotNil(t, pkt, "New packet should not be nil") {
		assert.Equal(t, "*packets1.WillMsg", reflect.TypeOf(pkt).String(), "Type should be WillMsg")
		assert.Equal(t, payload, pkt.WillMsg, "Bad WillMsg value")
	}
}

func TestWillMsgMarshal(t *testing.T) {
	assert := assert.New(t)
	buf := bytes.NewBuffer(nil)

	pkt1 := NewWillMsg([]byte("test-message"))
	if err := pkt1.Write(buf); err != nil {
		t.Fatal(err)
	}

	r := bytes.NewReader(buf.Bytes())
	pkt2, err := ReadPacket(r)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(pkt1, pkt2.(*WillMsg))
}
