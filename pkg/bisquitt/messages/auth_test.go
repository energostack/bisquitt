package messages

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthMarshal(t *testing.T) {
	assert := assert.New(t)
	buf := bytes.NewBuffer(nil)

	msg1 := NewAuthPlain("test-user", []byte("test-password"))
	if err := msg1.Write(buf); err != nil {
		t.Fatal(err)
	}

	r := bytes.NewReader(buf.Bytes())
	msg2, err := ReadPacket(r)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(msg1, msg2.(*AuthMessage))
}
