package packets1

import (
	"bytes"
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
	assert := assert.New(t)
	buf := bytes.NewBuffer(nil)

	pkt1 := NewDisconnect(75)
	err := pkt1.Write(buf)
	if err != nil {
		t.Fatal(err)
	}

	r := bytes.NewReader(buf.Bytes())
	pkt2, err := ReadPacket(r)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(pkt1, pkt2.(*Disconnect))
}
