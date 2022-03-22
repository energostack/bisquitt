package messages

import (
	"io"
)

const pingrespVarPartLength uint16 = 0

type PingrespMessage struct {
	Header
}

func NewPingrespMessage() *PingrespMessage {
	return &PingrespMessage{
		Header: *NewHeader(PINGRESP, pingrespVarPartLength),
	}
}

func (m *PingrespMessage) Write(w io.Writer) error {
	buf := m.Header.pack()

	_, err := buf.WriteTo(w)
	return err
}

func (m *PingrespMessage) Unpack(r io.Reader) error {
	return nil
}

func (m PingrespMessage) String() string {
	return "PINGRESP"
}
