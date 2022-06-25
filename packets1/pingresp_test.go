package packets1

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPingrespStruct(t *testing.T) {
	pkt := NewPingresp()

	if assert.NotNil(t, pkt, "New packet should not be nil") {
		assert.Equal(t, "*packets1.Pingresp", reflect.TypeOf(pkt).String(), "Type should be Pingresp")
		assert.Equal(t, uint16(2), pkt.PacketLength(), "Default Length should be 2")
	}
}

func TestPingrespMarshal(t *testing.T) {
	assert := assert.New(t)
	buf := bytes.NewBuffer(nil)

	pkt1 := NewPingresp()
	if err := pkt1.Write(buf); err != nil {
		t.Fatal(err)
	}

	r := bytes.NewReader(buf.Bytes())
	pkt2, err := ReadPacket(r)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(pkt1, pkt2.(*Pingresp))
}
