package messages

import (
	"fmt"
	"io"
)

const disconnectDurationLength uint16 = 2

type DisconnectMessage struct {
	Header
	Duration uint16
}

// NOTE: Packet length is initialized in this constructor and recomputed in m.Write().
func NewDisconnectMessage(duration uint16) *DisconnectMessage {
	m := &DisconnectMessage{
		Header:   *NewHeader(DISCONNECT, 0),
		Duration: duration,
	}
	m.computeLength()
	return m
}

func (m *DisconnectMessage) computeLength() {
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

func (m *DisconnectMessage) Write(w io.Writer) error {
	m.computeLength()

	buf := m.Header.pack()
	if m.VarPartLength() > 0 {
		buf.Write(encodeUint16(m.Duration))
	}

	_, err := buf.WriteTo(w)
	return err
}

func (m *DisconnectMessage) Unpack(r io.Reader) (err error) {
	if m.VarPartLength() == disconnectDurationLength {
		m.Duration, err = readUint16(r)
	} else {
		m.Duration = 0
	}
	return
}

func (m DisconnectMessage) String() string {
	return fmt.Sprintf(
		"DISCONNECT(Duration=%v)", m.Duration)
}
