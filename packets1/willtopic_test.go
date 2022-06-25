package packets1

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWillTopicStruct(t *testing.T) {
	qos := uint8(1)
	retain := false
	willTopic := "test-topic"
	pkt := NewWillTopic(willTopic, qos, retain)

	if assert.NotNil(t, pkt, "New packet should not be nil") {
		assert.Equal(t, "*packets1.WillTopic", reflect.TypeOf(pkt).String(), "Type should be WillTopic")
		assert.Equal(t, qos, pkt.QOS, "Bad QOS value")
		assert.Equal(t, retain, pkt.Retain, "Bad Retain flag value")
		assert.Equal(t, willTopic, pkt.WillTopic, "Bad WillTopic value")
	}
}

func TestWillTopicMarshal(t *testing.T) {
	assert := assert.New(t)
	buf := bytes.NewBuffer(nil)

	pkt1 := NewWillTopic("test-topic", 1, true)
	if err := pkt1.Write(buf); err != nil {
		t.Fatal(err)
	}

	r := bytes.NewReader(buf.Bytes())
	pkt2, err := ReadPacket(r)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(pkt1, pkt2.(*WillTopic))
}
