package packets1

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWillMsgUpdStruct(t *testing.T) {
	willMsg := []byte("test-msg")
	pkt := NewWillMsgUpd(willMsg)

	if assert.NotNil(t, pkt, "New packet should not be nil") {
		assert.Equal(t, "*packets1.WillMsgUpd", reflect.TypeOf(pkt).String(), "Type should be WillMsgUpd")
		assert.Equal(t, willMsg, pkt.WillMsg, "Bad WillMsg value")
	}
}

func TestWillMsgUpdMarshal(t *testing.T) {
	pkt1 := NewWillMsgUpd([]byte("test-message"))
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*WillMsgUpd))
}
