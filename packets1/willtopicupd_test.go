package packets1

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	pkts "github.com/energomonitor/bisquitt/packets"
)

func TestWillTopicUpdConstructor(t *testing.T) {
	assert := assert.New(t)

	qos := uint8(1)
	retain := false
	willTopic := "test-topic"
	pkt := NewWillTopicUpd(willTopic, qos, retain)

	if pkt == nil {
		t.Fatal("New packet should not be nil")
	}

	assert.Equal("*packets1.WillTopicUpd", reflect.TypeOf(pkt).String(), "Type should be WillTopicUpd")
	assert.Equal(qos, pkt.QOS, "Bad QOS value")
	assert.Equal(retain, pkt.Retain, "Bad Retain flag value")
	assert.Equal(willTopic, pkt.WillTopic, "Bad WillTopicUpd value")
}

func TestWillTopicUpdMarshal(t *testing.T) {
	pkt1 := NewWillTopicUpd("test-topic", 1, true)
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*WillTopicUpd))

	// Packet with a zero-length topic is a special case (does not contain Flags).
	pkt1 = NewWillTopicUpd("", 0, false)
	pkt2 = testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*WillTopicUpd))
}

func TestWillTopicUpdUnmarshalInvalid(t *testing.T) {
	assert := assert.New(t)

	// Packet too short.
	buff := bytes.NewBuffer([]byte{
		3,                       // Length
		byte(pkts.WILLTOPICUPD), // MsgType
		0,                       // Flags
		// Will Topic missing
	})
	_, err := ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad WILLTOPICUPD packet length")
	}
}

func TestWillTopicUpdStringer(t *testing.T) {
	pkt := NewWillTopicUpd("test-topic", 1, true)
	assert.Equal(t, `WILLTOPICUPD(WillTopic="test-topic", QOS=1, Retain=true)`, pkt.String())

	// Packet with a zero-length topic.
	pkt = NewWillTopicUpd("", 0, false)
	assert.Equal(t, `WILLTOPICUPD(WillTopic="")`, pkt.String())
}
