package packets1

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPingreqConstructor(t *testing.T) {
	assert := assert.New(t)

	clientID := []byte("test-client")
	pkt := NewPingreq(clientID)

	if pkt == nil {
		t.Fatal("New packet should not be nil")
	}

	assert.Equal("*packets1.Pingreq", reflect.TypeOf(pkt).String(), "Type should be Pingreq")
	assert.Equal(clientID, pkt.ClientID, "Bad ClientID value")
}

func TestPingreqMarshal(t *testing.T) {
	pkt1 := NewPingreq([]byte("test-client"))
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*Pingreq))
}

func TestPingreqStringer(t *testing.T) {
	pkt := NewPingreq([]byte("test-client"))
	assert.Equal(t, `PINGREQ(ClientID="test-client")`, pkt.String())
}
