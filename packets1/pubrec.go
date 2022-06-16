package packets1

import (
	"fmt"
	"io"
)

const pubrecVarPartLength uint16 = 2

type Pubrec struct {
	Header
	MessageIDProperty
}

func NewPubrec() *Pubrec {
	return &Pubrec{
		Header: *NewHeader(PUBREC, pubrecVarPartLength),
	}
}

func (m *Pubrec) Write(w io.Writer) error {
	buf := m.Header.pack()
	buf.Write(encodeUint16(m.messageID))

	_, err := buf.WriteTo(w)
	return err
}

func (m *Pubrec) Unpack(r io.Reader) (err error) {
	m.messageID, err = readUint16(r)
	return
}

func (m Pubrec) String() string {
	return fmt.Sprintf("PUBREC(MessageID=%d)", m.messageID)
}
