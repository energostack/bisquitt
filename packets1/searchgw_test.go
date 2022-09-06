package packets1

import (
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
	pkt1 := NewSearchGw(123)
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*SearchGw))
}
