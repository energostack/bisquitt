package packets1

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWillMsgUpdateStruct(t *testing.T) {
	willMsg := []byte("test-msg")
	pkt := NewWillMsgUpdate(willMsg)

	if assert.NotNil(t, pkt, "New packet should not be nil") {
		assert.Equal(t, "*packets1.WillMsgUpdate", reflect.TypeOf(pkt).String(), "Type should be WillMsgUpdate")
		assert.Equal(t, willMsg, pkt.WillMsg, "Bad WillMsg value")
	}
}

func TestWillMsgUpdateMarshal(t *testing.T) {
	assert := assert.New(t)
	buf := bytes.NewBuffer(nil)

	pkt1 := NewWillMsgUpdate([]byte("test-message"))
	if err := pkt1.Write(buf); err != nil {
		t.Fatal(err)
	}

	r := bytes.NewReader(buf.Bytes())
	pkt2, err := ReadPacket(r)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(pkt1, pkt2.(*WillMsgUpdate))
}
