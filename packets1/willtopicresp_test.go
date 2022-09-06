package packets1

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWillTopicRespStruct(t *testing.T) {
	pkt := NewWillTopicResp(RC_ACCEPTED)

	if assert.NotNil(t, pkt, "New packet should not be nil") {
		assert.Equal(t, "*packets1.WillTopicResp", reflect.TypeOf(pkt).String(), "Type should be WillTopicResp")
		assert.Equal(t, RC_ACCEPTED, pkt.ReturnCode, "ReturnCode should be RC_ACCEPTED")
	}
}

func TestWillTopicRespMarshal(t *testing.T) {
	pkt1 := NewWillTopicResp(RC_CONGESTION)
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*WillTopicResp))
}
