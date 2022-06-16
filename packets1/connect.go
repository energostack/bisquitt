package packets1

import (
	"fmt"
	"io"
)

const connectHeaderLength uint16 = 4

type Connect struct {
	Header
	CleanSession bool
	ClientID     []byte
	Duration     uint16
	ProtocolID   uint8
	Will         bool
}

// NOTE: Packet length is initialized in this constructor and recomputed in m.Write().
func NewConnect(clientID []byte, cleanSession bool, will bool, duration uint16) *Connect {
	m := &Connect{
		Header:       *NewHeader(CONNECT, 0),
		Will:         will,
		CleanSession: cleanSession,
		ProtocolID:   0x01,
		Duration:     duration,
		ClientID:     clientID,
	}
	m.computeLength()
	return m
}

func (m *Connect) computeLength() {
	clientIDLength := uint16(len(m.ClientID))
	m.Header.SetVarPartLength(connectHeaderLength + clientIDLength)
}

func (m *Connect) decodeFlags(b byte) {
	m.Will = (b & flagsWillBit) == flagsWillBit
	m.CleanSession = (b & flagsCleanSessionBit) == flagsCleanSessionBit
}

func (m *Connect) encodeFlags() byte {
	var b byte
	if m.Will {
		b |= flagsWillBit
	}
	if m.CleanSession {
		b |= flagsCleanSessionBit
	}
	return b
}

func (m *Connect) Write(w io.Writer) error {
	m.computeLength()

	buf := m.Header.pack()
	buf.WriteByte(m.encodeFlags())
	buf.WriteByte(m.ProtocolID)
	buf.Write(encodeUint16(m.Duration))
	buf.Write([]byte(m.ClientID))

	_, err := buf.WriteTo(w)
	return err
}

func (m *Connect) Unpack(r io.Reader) (err error) {
	var flagsByte uint8
	if flagsByte, err = readByte(r); err != nil {
		return
	}
	m.decodeFlags(flagsByte)

	if m.ProtocolID, err = readByte(r); err != nil {
		return
	}

	if m.Duration, err = readUint16(r); err != nil {
		return
	}

	m.ClientID = make([]byte, m.VarPartLength()-connectHeaderLength)
	_, err = io.ReadFull(r, m.ClientID)
	return
}

func (m Connect) String() string {
	return fmt.Sprintf(
		"CONNECT(ClientID=%#v, CleanSession=%t, Will=%t, Duration=%d)",

		string(m.ClientID), m.CleanSession, m.Will, m.Duration,
	)
}
