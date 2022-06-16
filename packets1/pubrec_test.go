package packets1

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPubrecStruct(t *testing.T) {
	msg := NewPubrec()

	if assert.NotNil(t, msg, "New packet should not be nil") {
		assert.Equal(t, "*packets1.Pubrec", reflect.TypeOf(msg).String(), "Type should be Pubrec")
		assert.Equal(t, uint16(4), msg.PacketLength(), "Default Length should be 4")
		assert.Equal(t, uint16(0), msg.MessageID(), "Default MessageID should be 0")
	}
}

func TestPubrecMarshal(t *testing.T) {
	assert := assert.New(t)
	buf := bytes.NewBuffer(nil)

	msg1 := NewPubrec()
	msg1.SetMessageID(12)
	if err := msg1.Write(buf); err != nil {
		t.Fatal(err)
	}

	r := bytes.NewReader(buf.Bytes())
	msg2, err := ReadPacket(r)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(msg1, msg2.(*Pubrec))
}
