package packets1

import (
	"bytes"
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
	assert := assert.New(t)
	buf := bytes.NewBuffer(nil)

	pkt1 := NewAdvertise(12, 123)
	if err := pkt1.Write(buf); err != nil {
		t.Fatal(err)
	}

	r := bytes.NewReader(buf.Bytes())
	pkt2, err := ReadPacket(r)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(pkt1, pkt2.(*Advertise))
}
