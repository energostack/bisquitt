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
	msg := NewGwInfoMessage(gatewayID, gatewayAddress)

	if assert.NotNil(t, msg, "New message should not be nil") {
		assert.Equal(t, "*packets1.GwInfoMessage", reflect.TypeOf(msg).String(), "Type should be GwInfoMessage")
		assert.Equal(t, gatewayID, msg.GatewayID, "Bad GatewayID value")
		assert.Equal(t, gatewayAddress, msg.GatewayAddress, "Bad GatewayAddress value")
	}
}

func TestGwInfoMarshal(t *testing.T) {
	assert := assert.New(t)
	buf := bytes.NewBuffer(nil)

	msg1 := NewGwInfoMessage(123, []byte("gateway-address"))
	if err := msg1.Write(buf); err != nil {
		t.Fatal(err)
	}

	r := bytes.NewReader(buf.Bytes())
	msg2, err := ReadPacket(r)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(msg1, msg2.(*GwInfoMessage))
}
