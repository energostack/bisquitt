package packets1

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	pkts "github.com/energomonitor/bisquitt/packets"
)

func TestWillMsgRespConstructor(t *testing.T) {
	assert := assert.New(t)

	pkt := NewWillMsgResp(RC_CONGESTION)

	if pkt == nil {
		t.Fatal("New packet should not be nil")
	}

	assert.Equal("*packets1.WillMsgResp", reflect.TypeOf(pkt).String(), "Type should be WillMsgResp")
	assert.Equal(RC_CONGESTION, pkt.ReturnCode, "ReturnCode should be RC_CONGESTION")
}

func TestWillMsgRespMarshal(t *testing.T) {
	pkt1 := NewWillMsgResp(RC_CONGESTION)
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*WillMsgResp))
}

func TestWillMsgRespUnmarshalInvalid(t *testing.T) {
	assert := assert.New(t)

	// Packet too short.
	buff := bytes.NewBuffer([]byte{
		2,                      // Length
		byte(pkts.WILLMSGRESP), // MsgType
		// Return Code missing
	})
	_, err := ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad WILLMSGRESP packet length")
	}

	// Packet too long.
	buff = bytes.NewBuffer([]byte{
		4,                      // Length
		byte(pkts.WILLMSGRESP), // MsgType
		0,                      // Return Code
		0,                      // junk
	})
	_, err = ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad WILLMSGRESP packet length")
	}
}

func TestWillMsgRespStringer(t *testing.T) {
	pkt := NewWillMsgResp(RC_CONGESTION)
	assert.Equal(t, "WILLMSGRESP(ReturnCode=1)", pkt.String())
}
