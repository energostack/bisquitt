package packets2

import (
	"bytes"
	"reflect"
	"testing"

	pkts "github.com/energomonitor/bisquitt/packets"
	"github.com/stretchr/testify/assert"
)

func TestPingrespConstructor(t *testing.T) {
	assert := assert.New(t)

	msgsRemaining := uint8(123)
	msgsRemainingPresent := true
	pkt := NewPingresp(msgsRemaining, msgsRemainingPresent)

	if pkt == nil {
		t.Fatal("New packet should not be nil")
	}

	assert.Equal("*packets2.Pingresp", reflect.TypeOf(pkt).String(), "Type should be Pingresp")
	assert.Equal(msgsRemaining, pkt.MessagesRemaining, "Bad MessagesRemaining value")
	assert.Equal(msgsRemainingPresent, pkt.MessagesRemainingPresent, "Bad MessagesRemainingPresent value")
	assert.Equal(uint16(3), pkt.PacketLength(), "Length should be 3")
}

func TestPingrespMarshal(t *testing.T) {
	// Packet without Messages Remaining
	pkt1 := NewPingresp(0, false)
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*Pingresp))

	// Packet with Messages Remaining
	pkt1 = NewPingresp(123, true)
	pkt2 = testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*Pingresp))
}

func TestPingrespUnmarshalInvalid(t *testing.T) {
	assert := assert.New(t)

	// Packet too long.
	buff := bytes.NewBuffer([]byte{
		6,                   // Length
		byte(pkts.PINGRESP), // Packet Type
		0,                   // Messages Remaining
		0,                   // junk
	})
	_, err := ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad PINGRESP2 packet length")
	}
}

func TestPingrespStringer(t *testing.T) {
	// Packet without Messages Remaining
	pkt := NewPingresp(0, false)
	assert.Equal(t, "PINGRESP2", pkt.String())

	// Packet with Messages Remaining
	pkt = NewPingresp(123, true)
	assert.Equal(t, "PINGRESP2(MsgsRemaining=123)", pkt.String())
}
