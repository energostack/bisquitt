package packets1

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPubrelStruct(t *testing.T) {
	pkt := NewPubrel()

	if assert.NotNil(t, pkt, "New packet should not be nil") {
		assert.Equal(t, "*packets1.Pubrel", reflect.TypeOf(pkt).String(), "Type should be Pubrel")
		assert.Equal(t, uint16(4), pkt.PacketLength(), "Default Length should be 4")
		assert.Equal(t, uint16(0), pkt.MessageID(), "Default MessageID should be 0")
	}
}

func TestPubrelMarshal(t *testing.T) {
	assert := assert.New(t)
	buf := bytes.NewBuffer(nil)

	pkt1 := NewPubrel()
	pkt1.SetMessageID(12)
	if err := pkt1.Write(buf); err != nil {
		t.Fatal(err)
	}

	r := bytes.NewReader(buf.Bytes())
	pkt2, err := ReadPacket(r)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(pkt1, pkt2.(*Pubrel))
}
