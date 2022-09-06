package packets1

import (
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
	pkt1 := NewWillMsg([]byte("test-message"))
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*WillMsg))
}
