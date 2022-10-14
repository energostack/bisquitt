package packets1

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPingrespConstructor(t *testing.T) {
	pkt := NewPingresp()

	if pkt == nil {
		t.Fatal("New packet should not be nil")
	}

	assert.Equal(t, "*packets1.Pingresp", reflect.TypeOf(pkt).String(), "Type should be Pingresp")
	assert.Equal(t, uint16(2), pkt.PacketLength(), "Default Length should be 2")
}

func TestPingrespMarshal(t *testing.T) {
	pkt1 := NewPingresp()
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*Pingresp))
}

func TestPingrespStringer(t *testing.T) {
	pkt := NewPingresp()
	assert.Equal(t, "PINGRESP", pkt.String())
}
