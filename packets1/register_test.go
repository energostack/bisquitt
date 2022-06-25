package packets1

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegisterStruct(t *testing.T) {
	topicID := uint16(123)
	topic := "test-topic"
	pkt := NewRegister(topicID, topic)

	if assert.NotNil(t, pkt, "New packet should not be nil") {
		assert.Equal(t, "*packets1.Register", reflect.TypeOf(pkt).String(), "Type should be Register")
		assert.Equal(t, topicID, pkt.TopicID, "Bad TopicID value")
		assert.Equal(t, uint16(0), pkt.MessageID(), "Default MessageID should be 0")
		assert.Equal(t, topic, pkt.TopicName, "Bad TopicName value")
	}
}

func TestRegisterMarshal(t *testing.T) {
	assert := assert.New(t)
	buf := bytes.NewBuffer(nil)

	pkt1 := NewRegister(123, "test-topic")
	pkt1.SetMessageID(12)
	if err := pkt1.Write(buf); err != nil {
		t.Fatal(err)
	}

	r := bytes.NewReader(buf.Bytes())
	pkt2, err := ReadPacket(r)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(pkt1, pkt2.(*Register))
}
