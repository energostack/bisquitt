package packets1

import (
	"bytes"
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
	assert := assert.New(t)
	buf := bytes.NewBuffer(nil)

	pkt1 := NewPingreq([]byte("test-client"))
	if err := pkt1.Write(buf); err != nil {
		t.Fatal(err)
	}

	r := bytes.NewReader(buf.Bytes())
	pkt2, err := ReadPacket(r)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(pkt1, pkt2.(*Pingreq))
}
