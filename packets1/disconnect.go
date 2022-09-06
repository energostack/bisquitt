package packets1

import (
	"encoding/binary"
	"fmt"

	pkts "github.com/energomonitor/bisquitt/packets"
)

const disconnectDurationLength uint16 = 2

type Disconnect struct {
	pkts.Header
	Duration uint16
}

// NOTE: Packet length is initialized in this constructor and recomputed in m.Write().
func NewDisconnect(duration uint16) *Disconnect {
	p := &Disconnect{
		Header:   *pkts.NewHeader(pkts.DISCONNECT, 0),
		Duration: duration,
	}
	p.computeLength()
	return p
}

func (p *Disconnect) computeLength() {
	// Duration: contains the value of the sleep timer; this field is
	// optional and is included by a “sleeping” client that wants to go the
	// “asleep” state
	// [MQTT-SN specification v. 1.2, chapter 5.4.21 DISCONNECT]
	if p.Duration == 0 {
		p.Header.SetVarPartLength(0)
	} else {
		p.Header.SetVarPartLength(disconnectDurationLength)
	}
}

func (p *Disconnect) Pack() ([]byte, error) {
	p.computeLength()
	buf := p.Header.PackToBuffer()

	if p.VarPartLength() > 0 {
		_, _ = buf.Write(pkts.EncodeUint16(p.Duration))
	}

	return buf.Bytes(), nil
}

func (p *Disconnect) Unpack(buf []byte) error {
	switch len(buf) {
	case int(disconnectDurationLength):
		p.Duration = binary.BigEndian.Uint16(buf)
	case 0:
		p.Duration = 0
	default:
		return fmt.Errorf("bad DISCONNECT packet length: expected 0 or 2, got %d",
			len(buf))
	}

	return nil
}

func (p Disconnect) String() string {
	return fmt.Sprintf(
		"DISCONNECT(Duration=%v)", p.Duration)
}
