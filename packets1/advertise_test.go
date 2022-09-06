package packets1

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAdvertiseStruct(t *testing.T) {
	gatewayID := uint8(12)
	duration := uint16(123)
	pkt := NewAdvertise(gatewayID, duration)

	if assert.NotNil(t, pkt, "New packet should not be nil") {
		assert.Equal(t, "*packets1.Advertise", reflect.TypeOf(pkt).String(), "Type should be Advertise")
		assert.Equal(t, gatewayID, pkt.GatewayID, "Bad GatewayID")
		assert.Equal(t, duration, pkt.Duration, "Bad Duration value")
		assert.Equal(t, uint16(5), pkt.PacketLength(), "Length should be 5")
	}
}

func TestAdvertiseMarshal(t *testing.T) {
	pkt1 := NewAdvertise(12, 123)
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*Advertise))
}
