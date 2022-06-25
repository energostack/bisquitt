package packets1

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthMarshal(t *testing.T) {
	assert := assert.New(t)
	buf := bytes.NewBuffer(nil)

	pkt1 := NewAuthPlain("test-user", []byte("test-password"))
	if err := pkt1.Write(buf); err != nil {
		t.Fatal(err)
	}

	r := bytes.NewReader(buf.Bytes())
	pkt2, err := ReadPacket(r)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(pkt1, pkt2.(*Auth))
}
