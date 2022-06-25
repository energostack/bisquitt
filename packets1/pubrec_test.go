package packets1

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPubrecStruct(t *testing.T) {
	pkt := NewPubrec()

	if assert.NotNil(t, pkt, "New packet should not be nil") {
		assert.Equal(t, "*packets1.Pubrec", reflect.TypeOf(pkt).String(), "Type should be Pubrec")
		assert.Equal(t, uint16(4), pkt.PacketLength(), "Default Length should be 4")
		assert.Equal(t, uint16(0), pkt.MessageID(), "Default MessageID should be 0")
	}
}

func TestPubrecMarshal(t *testing.T) {
	assert := assert.New(t)
	buf := bytes.NewBuffer(nil)

	pkt1 := NewPubrec()
	pkt1.SetMessageID(12)
	if err := pkt1.Write(buf); err != nil {
		t.Fatal(err)
	}

	r := bytes.NewReader(buf.Bytes())
	pkt2, err := ReadPacket(r)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(pkt1, pkt2.(*Pubrec))
}
