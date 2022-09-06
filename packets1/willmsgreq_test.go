package packets1

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWillMsgReqStruct(t *testing.T) {
	pkt := NewWillMsgReq()

	if assert.NotNil(t, pkt, "New packet should not be nil") {
		assert.Equal(t, "*packets1.WillMsgReq", reflect.TypeOf(pkt).String(), "Type should be WillMsgReq")
	}
}

func TestWillMsgReqMarshal(t *testing.T) {
	pkt1 := NewWillMsgReq()
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*WillMsgReq))
}
