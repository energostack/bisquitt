package packets1

import (
	"fmt"
	"io"
)

const pubrelVarPartLength uint16 = 2

type Pubrel struct {
	Header
	MessageIDProperty
}

func NewPubrel() *Pubrel {
	return &Pubrel{
		Header: *NewHeader(PUBREL, pubrelVarPartLength),
	}
}

func (m *Pubrel) Write(w io.Writer) error {
	buf := m.Header.pack()
	buf.Write(encodeUint16(m.messageID))

	_, err := buf.WriteTo(w)
	return err
}

func (m *Pubrel) Unpack(r io.Reader) (err error) {
	m.messageID, err = readUint16(r)
	return
}

func (m Pubrel) String() string {
	return fmt.Sprintf("PUBREL(MessageID=%d)", m.messageID)
}
