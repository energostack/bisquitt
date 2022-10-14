package packets1

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	pkts "github.com/energomonitor/bisquitt/packets"
)

func TestWillTopicReqConstructor(t *testing.T) {
	pkt := NewWillTopicReq()

	if pkt == nil {
		t.Fatal("New packet should not be nil")
	}

	assert.Equal(t, "*packets1.WillTopicReq", reflect.TypeOf(pkt).String(), "Type should be WillTopicReq")
}

func TestWillTopicReqMarshal(t *testing.T) {
	pkt1 := NewWillTopicReq()
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*WillTopicReq))
}

func TestWillTopicReqUnmarshalInvalid(t *testing.T) {
	assert := assert.New(t)

	// Packet too long.
	buff := bytes.NewBuffer([]byte{
		3,                       // Length
		byte(pkts.WILLTOPICREQ), // MsgType
		0,                       // junk
	})
	_, err := ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad WILLTOPICREQ packet length")
	}
}

func TestWillTopicReqStringer(t *testing.T) {
	pkt := NewWillTopicReq()
	assert.Equal(t, "WILLTOPICREQ", pkt.String())
}
