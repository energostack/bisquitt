package packets1

import (
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
	pkt1 := NewWillMsgUpdate([]byte("test-message"))
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*WillMsgUpdate))
}
