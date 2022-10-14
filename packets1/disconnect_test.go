package packets1

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	pkts "github.com/energomonitor/bisquitt/packets"
)

func TestDisconnectConstructor(t *testing.T) {
	assert := assert.New(t)

	duration := uint16(123)
	pkt := NewDisconnect(duration)

	if pkt == nil {
		t.Fatal("New packet should not be nil")
	}

	assert.Equal("*packets1.Disconnect", reflect.TypeOf(pkt).String(), "Type should be Disconnect")
	assert.Equal(duration, pkt.Duration, "Bad Duration value")
}

func TestDisconnectMarshal(t *testing.T) {
	pkt1 := NewDisconnect(75)
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*Disconnect))
}

func TestDisconnectUnmarshal(t *testing.T) {
	assert := assert.New(t)

	// Packet without Duration is correct (length=2).
	buff := bytes.NewBuffer([]byte{
		2,                     // Length
		byte(pkts.DISCONNECT), // MsgType
	})
	_, err := ReadPacket(buff)
	assert.Nil(err)

	// Packet too short (length=3).
	buff = bytes.NewBuffer([]byte{
		3,                     // Length
		byte(pkts.DISCONNECT), // MsgType
		0,                     // Duration MSB
		// Duration LSB missing
	})
	_, err = ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad DISCONNECT packet length")
	}

	// Packet with Duration is correct (length=4).
	buff = bytes.NewBuffer([]byte{
		4,                     // Length
		byte(pkts.DISCONNECT), // MsgType
		0, 1,                  // Duration
	})
	_, err = ReadPacket(buff)
	assert.Nil(err)

	// Packet too long (length=5).
	buff = bytes.NewBuffer([]byte{
		5,                     // Length
		byte(pkts.DISCONNECT), // MsgType
		0,                     // Duration MSB
		1,                     // Duration LSB
		0,                     // junk
	})
	_, err = ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad DISCONNECT packet length")
	}
}

func TestDisconnectStringer(t *testing.T) {
	// Packet without Duration.
	pkt := NewDisconnect(0)
	assert.Equal(t, "DISCONNECT(Duration=0)", pkt.String())

	// Packet with Duration.
	pkt = NewDisconnect(1234)
	assert.Equal(t, "DISCONNECT(Duration=1234)", pkt.String())
}
