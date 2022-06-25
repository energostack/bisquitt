package packets1

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSearchGwStruct(t *testing.T) {
	radius := uint8(123)
	pkt := NewSearchGw(radius)

	if assert.NotNil(t, pkt, "New packet should not be nil") {
		assert.Equal(t, "*packets1.SearchGw", reflect.TypeOf(pkt).String(), "Type should be SearchGw")
		assert.Equal(t, uint16(3), pkt.PacketLength(), "Default Length should be 3")
		assert.Equal(t, radius, pkt.Radius, "Bad Radius value")
	}
}

func TestSearchGwMarshal(t *testing.T) {
	assert := assert.New(t)
	buf := bytes.NewBuffer(nil)

	pkt1 := NewSearchGw(123)
	if err := pkt1.Write(buf); err != nil {
		t.Fatal(err)
	}

	r := bytes.NewReader(buf.Bytes())
	pkt2, err := ReadPacket(r)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(pkt1, pkt2.(*SearchGw))
}
