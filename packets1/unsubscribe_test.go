package packets1

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnsubscribeStruct(t *testing.T) {
	topicID := uint16(12)
	topicIDType := TIT_REGISTERED
	topicName := []byte("test-topic")
	pkt := NewUnsubscribe(topicID, topicIDType, topicName)

	if assert.NotNil(t, pkt, "New packet should not be nil") {
		assert.Equal(t, "*packets1.Unsubscribe", reflect.TypeOf(pkt).String(), "Type should be Unsubscribe")
		assert.Equal(t, topicIDType, pkt.TopicIDType, "Bad TopicIDType value")
		assert.Equal(t, topicID, pkt.TopicID, "Bad TopicID value")
		assert.Equal(t, uint16(0), pkt.MessageID(), "Default MessageID should be 0")
		assert.Equal(t, topicName, pkt.TopicName, "Bad Topicname value")
	}
}

func TestUnsubscribeMarshalString(t *testing.T) {
	assert := assert.New(t)
	buf := bytes.NewBuffer(nil)

	pkt1 := NewUnsubscribe(0, TIT_STRING, []byte("test-topic"))
	pkt1.SetMessageID(12)
	if err := pkt1.Write(buf); err != nil {
		t.Fatal(err)
	}

	r := bytes.NewReader(buf.Bytes())
	pkt2, err := ReadPacket(r)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(pkt1, pkt2.(*Unsubscribe))
}

func TestUnsubscribeMarshalShort(t *testing.T) {
	assert := assert.New(t)
	buf := bytes.NewBuffer(nil)

	pkt1 := NewUnsubscribe(123, TIT_SHORT, nil)
	pkt1.SetMessageID(12)
	if err := pkt1.Write(buf); err != nil {
		t.Fatal(err)
	}

	r := bytes.NewReader(buf.Bytes())
	pkt2, err := ReadPacket(r)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(pkt1, pkt2.(*Unsubscribe))
}
