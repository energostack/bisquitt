package packets1

import (
	"bytes"
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
	assert := assert.New(t)
	buf := bytes.NewBuffer(nil)

	pkt1 := NewGwInfo(123, []byte("gateway-address"))
	if err := pkt1.Write(buf); err != nil {
		t.Fatal(err)
	}

	r := bytes.NewReader(buf.Bytes())
	pkt2, err := ReadPacket(r)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(pkt1, pkt2.(*GwInfo))
}
