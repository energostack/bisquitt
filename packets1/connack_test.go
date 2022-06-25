package packets1

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConnackStruct(t *testing.T) {
	pkt := NewConnack(RC_ACCEPTED)

	if assert.NotNil(t, pkt, "New packet should not be nil") {
		assert.Equal(t, "*packets1.Connack", reflect.TypeOf(pkt).String(), "Type should be Connack")
		assert.Equal(t, RC_ACCEPTED, pkt.ReturnCode, "ReturnCode should be RC_ACCEPTED")
		assert.Equal(t, uint16(3), pkt.PacketLength(), "Length should be 3")
	}
}

func TestConnackMarshal(t *testing.T) {
	assert := assert.New(t)
	buf := bytes.NewBuffer(nil)

	pkt1 := NewConnack(RC_CONGESTION)
	if err := pkt1.Write(buf); err != nil {
		t.Fatal(err)
	}

	r := bytes.NewReader(buf.Bytes())
	pkt2, err := ReadPacket(r)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(pkt1, pkt2.(*Connack))
}
