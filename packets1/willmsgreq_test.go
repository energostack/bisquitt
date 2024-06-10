package packets1

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	pkts "github.com/energostack/bisquitt/packets"
)

func TestWillMsgReqConstructor(t *testing.T) {
	pkt := NewWillMsgReq()

	if pkt == nil {
		t.Fatal("New packet should not be nil")
	}

	assert.Equal(t, "*packets1.WillMsgReq", reflect.TypeOf(pkt).String(), "Type should be WillMsgReq")
}

func TestWillMsgReqMarshal(t *testing.T) {
	pkt1 := NewWillMsgReq()
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*WillMsgReq))
}

func TestWillMsgReqUnmarshalInvalid(t *testing.T) {
	assert := assert.New(t)

	// Packet too long.
	buff := bytes.NewBuffer([]byte{
		3,                     // Length
		byte(pkts.WILLMSGREQ), // MsgType
		0,                     // junk
	})
	_, err := ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad WILLMSGREQ packet length")
	}
}

func TestWillMsgReqStringer(t *testing.T) {
	pkt := NewWillMsgReq()
	assert.Equal(t, "WILLMSGREQ", pkt.String())
}
