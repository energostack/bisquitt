package packets1

import (
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
	pkt1 := NewPubrec()
	pkt1.SetMessageID(12)
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*Pubrec))
}
