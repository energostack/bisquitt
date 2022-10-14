package packets1

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	pkts "github.com/energomonitor/bisquitt/packets"
)

func TestGwInfoConstructor(t *testing.T) {
	assert := assert.New(t)

	gatewayID := uint8(123)
	gatewayAddress := []byte("test-gw")
	pkt := NewGwInfo(gatewayID, gatewayAddress)

	if pkt == nil {
		t.Fatal("New packet should not be nil")
	}

	assert.Equal("*packets1.GwInfo", reflect.TypeOf(pkt).String(), "Type should be GwInfo")
	assert.Equal(gatewayID, pkt.GatewayID, "Bad GatewayID value")
	assert.Equal(gatewayAddress, pkt.GatewayAddress, "Bad GatewayAddress value")
}

func TestGwInfoMarshal(t *testing.T) {
	pkt1 := NewGwInfo(123, []byte("gateway-address"))
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*GwInfo))
}

func TestGwInfoUnmarshalInvalid(t *testing.T) {
	// Packet too short.
	buff := bytes.NewBuffer([]byte{
		2,                 // Length
		byte(pkts.GWINFO), // MsgType
		// Gateway ID missing
	})
	_, err := ReadPacket(buff)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "bad GWINFO packet length")
	}
}

func TestGwInfoStringer(t *testing.T) {
	pkt := NewGwInfo(123, []byte("gateway-address"))
	assert.Equal(t, `GWINFO(GatewayID=123,GatewayAddress="gateway-address")`, pkt.String())
}
