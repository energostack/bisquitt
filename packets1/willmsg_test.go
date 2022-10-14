package packets1

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWillMsgConstructor(t *testing.T) {
	assert := assert.New(t)

	data := []byte("test-data")
	pkt := NewWillMsg(data)

	if pkt == nil {
		t.Fatal("New packet should not be nil")
	}

	assert.Equal("*packets1.WillMsg", reflect.TypeOf(pkt).String(), "Type should be WillMsg")
	assert.Equal(data, pkt.WillMsg, "Bad WillMsg value")
}

func TestWillMsgMarshal(t *testing.T) {
	pkt1 := NewWillMsg([]byte("test-message"))
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*WillMsg))
}

func TestWillMsgStringer(t *testing.T) {
	pkt := NewWillMsg([]byte("test-data"))
	assert.Equal(t, `WILLMSG(WillMsg="test-data")`, pkt.String())
}
