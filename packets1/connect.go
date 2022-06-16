package packets1

import (
	"fmt"
	"io"
)

const connectHeaderLength uint16 = 4

type ConnectMessage struct {
	Header
	CleanSession bool
	ClientID     []byte
	Duration     uint16
	ProtocolID   uint8
	Will         bool
}

// NOTE: Packet length is initialized in this constructor and recomputed in m.Write().
func NewConnectMessage(clientID []byte, cleanSession bool, will bool, duration uint16) *ConnectMessage {
	m := &ConnectMessage{
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

func (m *ConnectMessage) computeLength() {
	clientIDLength := uint16(len(m.ClientID))
	m.Header.SetVarPartLength(connectHeaderLength + clientIDLength)
}

func (m *ConnectMessage) decodeFlags(b byte) {
	m.Will = (b & flagsWillBit) == flagsWillBit
	m.CleanSession = (b & flagsCleanSessionBit) == flagsCleanSessionBit
}

func (m *ConnectMessage) encodeFlags() byte {
	var b byte
	if m.Will {
		b |= flagsWillBit
	}
	if m.CleanSession {
		b |= flagsCleanSessionBit
	}
	return b
}

func (m *ConnectMessage) Write(w io.Writer) error {
	m.computeLength()

	buf := m.Header.pack()
	buf.WriteByte(m.encodeFlags())
	buf.WriteByte(m.ProtocolID)
	buf.Write(encodeUint16(m.Duration))
	buf.Write([]byte(m.ClientID))

	_, err := buf.WriteTo(w)
	return err
}

func (m *ConnectMessage) Unpack(r io.Reader) (err error) {
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

func (m ConnectMessage) String() string {
	return fmt.Sprintf(
		"CONNECT(ClientID=%#v, CleanSession=%t, Will=%t, Duration=%d)",

		string(m.ClientID), m.CleanSession, m.Will, m.Duration,
	)
}
