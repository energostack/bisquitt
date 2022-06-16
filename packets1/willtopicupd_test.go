package packets1

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWillTopicUpdateStruct(t *testing.T) {
	willTopic := []byte("test-topic")
	qos := uint8(1)
	retain := true
	msg := NewWillTopicUpdate(willTopic, qos, retain)

	if assert.NotNil(t, msg, "New packet should not be nil") {
		assert.Equal(t, "*packets1.WillTopicUpdate", reflect.TypeOf(msg).String(), "Type should be WillTopicUpdate")
		assert.Equal(t, qos, msg.QOS, "Bad QOS value")
		assert.Equal(t, retain, msg.Retain, "Bad Retain flag value")
		assert.Equal(t, willTopic, msg.WillTopic, "Bad WillTopic value")
	}

}

func TestWillTopicUpdateMarshal(t *testing.T) {
	assert := assert.New(t)
	buf := bytes.NewBuffer(nil)

	msg1 := NewWillTopicUpdate([]byte("test-topic"), 1, true)
	if err := msg1.Write(buf); err != nil {
		t.Fatal(err)
	}

	r := bytes.NewReader(buf.Bytes())
	msg2, err := ReadPacket(r)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(msg1, msg2.(*WillTopicUpdate))
}
