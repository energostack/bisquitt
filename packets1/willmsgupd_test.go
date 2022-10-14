package packets1

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWillMsgUpdConstructor(t *testing.T) {
	assert := assert.New(t)

	data := []byte("test-data")
	pkt := NewWillMsgUpd(data)

	if pkt == nil {
		t.Fatal("New packet should not be nil")
	}

	assert.Equal("*packets1.WillMsgUpd", reflect.TypeOf(pkt).String(), "Type should be WillMsgUpd")
	assert.Equal(data, pkt.WillMsg, "Bad WillMsg value")
}

func TestWillMsgUpdMarshal(t *testing.T) {
	pkt1 := NewWillMsgUpd([]byte("test-message"))
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*WillMsgUpd))
}

func TestWillMsgUpdStringer(t *testing.T) {
	pkt := NewWillMsgUpd([]byte("test-data"))
	assert.Equal(t, `WILLMSGUPD(WillMsg="test-data")`, pkt.String())
}
