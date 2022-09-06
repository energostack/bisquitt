package packets1

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPingreqStruct(t *testing.T) {
	clientID := []byte("test-client")
	pkt := NewPingreq(clientID)

	if assert.NotNil(t, pkt, "New packet should not be nil") {
		assert.Equal(t, "*packets1.Pingreq", reflect.TypeOf(pkt).String(), "Type should be Pingreq")
		assert.Equal(t, clientID, pkt.ClientID, "Bad ClientID value")
	}
}

func TestPingreqMarshal(t *testing.T) {
	pkt1 := NewPingreq([]byte("test-client"))
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*Pingreq))
}
