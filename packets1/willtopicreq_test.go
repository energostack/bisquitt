package packets1

import (
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
	pkt1 := NewWillTopicReq()
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*WillTopicReq))
}
