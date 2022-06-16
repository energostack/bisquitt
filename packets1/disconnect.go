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
	m := &Disconnect{
		Header:   *NewHeader(DISCONNECT, 0),
		Duration: duration,
	}
	m.computeLength()
	return m
}

func (m *Disconnect) computeLength() {
	// Duration: contains the value of the sleep timer; this field is
	// optional and is included by a “sleeping” client that wants to go the
	// “asleep” state
	// [MQTT-SN specification v. 1.2, chapter 5.4.21 DISCONNECT]
	if m.Duration == 0 {
		m.Header.SetVarPartLength(0)
	} else {
		m.Header.SetVarPartLength(disconnectDurationLength)
	}
}

func (m *Disconnect) Write(w io.Writer) error {
	m.computeLength()

	buf := m.Header.pack()
	if m.VarPartLength() > 0 {
		buf.Write(encodeUint16(m.Duration))
	}

	_, err := buf.WriteTo(w)
	return err
}

func (m *Disconnect) Unpack(r io.Reader) (err error) {
	if m.VarPartLength() == disconnectDurationLength {
		m.Duration, err = readUint16(r)
	} else {
		m.Duration = 0
	}
	return
}

func (m Disconnect) String() string {
	return fmt.Sprintf(
		"DISCONNECT(Duration=%v)", m.Duration)
}
