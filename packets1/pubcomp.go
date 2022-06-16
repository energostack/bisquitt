package packets1

import (
	"fmt"
	"io"
)

const pubcompVarPartLength uint16 = 2

type Pubcomp struct {
	Header
	MessageIDProperty
}

func NewPubcomp() *Pubcomp {
	return &Pubcomp{
		Header: *NewHeader(PUBCOMP, pubcompVarPartLength),
	}
}

func (m *Pubcomp) Write(w io.Writer) error {
	buf := m.Header.pack()
	buf.Write(encodeUint16(m.messageID))

	_, err := buf.WriteTo(w)
	return err
}

func (m *Pubcomp) Unpack(r io.Reader) (err error) {
	m.messageID, err = readUint16(r)
	return
}

func (m Pubcomp) String() string {
	return fmt.Sprintf("PUBCOMP(MessageID=%d)", m.messageID)
}
