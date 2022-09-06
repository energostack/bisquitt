package packets1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthMarshal(t *testing.T) {
	pkt1 := NewAuthPlain("test-user", []byte("test-password"))
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*Auth))
}
