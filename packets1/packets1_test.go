package packets1

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	pkts "github.com/energomonitor/bisquitt/packets"
)

func testPacketMarshal(t *testing.T, pkt1 pkts.Packet) pkts.Packet {
	buf, err := pkt1.Pack()
	if err != nil {
		t.Fatal(err)
	}

	r := bytes.NewReader(buf)
	pkt2, err := ReadPacket(r)
	if err != nil {
		t.Fatal(err)
	}

	return pkt2
}

func TestUnmarshalShortPacket(t *testing.T) {
	buff := bytes.NewBuffer([]byte{
		1, // Length
		// MsgType missing
	})
	_, err := ReadPacket(buff)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "bad packet length")
	}
}

func TestUnmarshalInvalidPacketType(t *testing.T) {
	buff := bytes.NewBuffer([]byte{
		2,          // Length
		byte(0x19), // invalid MsgType
	})
	_, err := ReadPacket(buff)
	if assert.Error(t, err) {
		assert.Equal(t, err.Error(), "invalid MQTT-SN packet type")
	}
}
