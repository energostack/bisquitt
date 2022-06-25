package packets1

import (
	"bytes"
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
	assert := assert.New(t)
	buf := bytes.NewBuffer(nil)

	pkt1 := NewWillMsgResp(RC_CONGESTION)
	if err := pkt1.Write(buf); err != nil {
		t.Fatal(err)
	}

	r := bytes.NewReader(buf.Bytes())
	pkt2, err := ReadPacket(r)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(pkt1, pkt2.(*WillMsgResp))
}
