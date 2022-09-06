package packets1

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDisconnectStruct(t *testing.T) {
	duration := uint16(123)
	pkt := NewDisconnect(duration)

	if assert.NotNil(t, pkt, "New packet should not be nil") {
		assert.Equal(t, "*packets1.Disconnect", reflect.TypeOf(pkt).String(), "Type should be Disconnect")
		assert.Equal(t, duration, pkt.Duration, "Bad Duration value")
	}
}

func TestDisconnectMarshal(t *testing.T) {
	pkt1 := NewDisconnect(75)
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*Disconnect))
}
