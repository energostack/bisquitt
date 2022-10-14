package packets1

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	pkts "github.com/energomonitor/bisquitt/packets"
)

func TestWillTopicConstructor(t *testing.T) {
	assert := assert.New(t)

	qos := uint8(1)
	retain := false
	willTopic := "test-topic"
	pkt := NewWillTopic(willTopic, qos, retain)

	if pkt == nil {
		t.Fatal("New packet should not be nil")
	}

	assert.Equal("*packets1.WillTopic", reflect.TypeOf(pkt).String(), "Type should be WillTopic")
	assert.Equal(qos, pkt.QOS, "Bad QOS value")
	assert.Equal(retain, pkt.Retain, "Bad Retain flag value")
	assert.Equal(willTopic, pkt.WillTopic, "Bad WillTopic value")
}

func TestWillTopicMarshal(t *testing.T) {
	pkt1 := NewWillTopic("test-topic", 1, true)
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*WillTopic))

	// Packet with zero-length topic is a special case (does not contain Flags).
	pkt1 = NewWillTopic("", 0, false)
	pkt2 = testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*WillTopic))
}

func TestWillTopicUnmarshalInvalid(t *testing.T) {
	assert := assert.New(t)

	// Packet too short.
	buff := bytes.NewBuffer([]byte{
		3,                    // Length
		byte(pkts.WILLTOPIC), // MsgType
		0,                    // Flags
		// Will Topic missing
	})
	_, err := ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad WILLTOPIC packet length")
	}
}

func TestWillTopicStringer(t *testing.T) {
	pkt := NewWillTopic("test-topic", 1, true)
	assert.Equal(t, `WILLTOPIC(WillTopic="test-topic", QOS=1, Retain=true)`, pkt.String())

	// Packet with zero-length topic.
	pkt = NewWillTopic("", 0, false)
	assert.Equal(t, `WILLTOPIC(WillTopic="")`, pkt.String())
}
