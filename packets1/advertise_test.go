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
	msg := NewAdvertise(gatewayID, duration)

	if assert.NotNil(t, msg, "New packet should not be nil") {
		assert.Equal(t, "*packets1.Advertise", reflect.TypeOf(msg).String(), "Type should be Advertise")
		assert.Equal(t, gatewayID, msg.GatewayID, "Bad GatewayID")
		assert.Equal(t, duration, msg.Duration, "Bad Duration value")
		assert.Equal(t, uint16(5), msg.MessageLength(), "Length should be 5")
	}
}

func TestAdvertiseMarshal(t *testing.T) {
	assert := assert.New(t)
	buf := bytes.NewBuffer(nil)

	msg1 := NewAdvertise(12, 123)
	if err := msg1.Write(buf); err != nil {
		t.Fatal(err)
	}

	r := bytes.NewReader(buf.Bytes())
	msg2, err := ReadPacket(r)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(msg1, msg2.(*Advertise))
}
