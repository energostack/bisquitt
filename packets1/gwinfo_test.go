package packets1

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGwInfoStruct(t *testing.T) {
	gatewayID := uint8(123)
	gatewayAddress := []byte("test-gw")
	pkt := NewGwInfo(gatewayID, gatewayAddress)

	if assert.NotNil(t, pkt, "New packet should not be nil") {
		assert.Equal(t, "*packets1.GwInfo", reflect.TypeOf(pkt).String(), "Type should be GwInfo")
		assert.Equal(t, gatewayID, pkt.GatewayID, "Bad GatewayID value")
		assert.Equal(t, gatewayAddress, pkt.GatewayAddress, "Bad GatewayAddress value")
	}
}

func TestGwInfoMarshal(t *testing.T) {
	pkt1 := NewGwInfo(123, []byte("gateway-address"))
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*GwInfo))
}
