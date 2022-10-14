package packets1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMessageID(t *testing.T) {
	msgID := uint16(1234)

	pkt1 := NewPubrec()
	pkt1.SetMessageID(msgID)
	assert.Equal(t, msgID, pkt1.MessageID())

	pkt2 := NewPubrel()
	pkt2.CopyMessageID(pkt1)
	assert.Equal(t, msgID, pkt2.MessageID())
}
