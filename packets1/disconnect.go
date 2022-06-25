package packets1

import (
	"fmt"
	"io"
)

const disconnectDurationLength uint16 = 2

type Disconnect struct {
	Header
	Duration uint16
}

// NOTE: Packet length is initialized in this constructor and recomputed in m.Write().
func NewDisconnect(duration uint16) *Disconnect {
	p := &Disconnect{
		Header:   *NewHeader(DISCONNECT, 0),
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

func (p *Disconnect) Write(w io.Writer) error {
	p.computeLength()

	buf := p.Header.pack()
	if p.VarPartLength() > 0 {
		buf.Write(encodeUint16(p.Duration))
	}

	_, err := buf.WriteTo(w)
	return err
}

func (p *Disconnect) Unpack(r io.Reader) (err error) {
	if p.VarPartLength() == disconnectDurationLength {
		p.Duration, err = readUint16(r)
	} else {
		p.Duration = 0
	}
	return
}

func (p Disconnect) String() string {
	return fmt.Sprintf(
		"DISCONNECT(Duration=%v)", p.Duration)
}
