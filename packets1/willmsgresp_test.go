package packets1

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWillMsgRespStruct(t *testing.T) {
	pkt := NewWillMsgResp(RC_ACCEPTED)

	if assert.NotNil(t, pkt, "New packet should not be nil") {
		assert.Equal(t, "*packets1.WillMsgResp", reflect.TypeOf(pkt).String(), "Type should be WillMsgResp")
		assert.Equal(t, RC_ACCEPTED, pkt.ReturnCode, "Default ReturnCode should be RC_ACCEPTED")
	}
}

func TestWillMsgRespMarshal(t *testing.T) {
	pkt1 := NewWillMsgResp(RC_CONGESTION)
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*WillMsgResp))
}
