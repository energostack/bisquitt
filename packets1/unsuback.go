package packets1

import (
	"fmt"
	"io"
)

const unsubackVarPartLength uint16 = 2

type Unsuback struct {
	Header
	MessageIDProperty
}

func NewUnsuback() *Unsuback {
	return &Unsuback{
		Header: *NewHeader(UNSUBACK, unsubackVarPartLength),
	}
}

func (m *Unsuback) Write(w io.Writer) error {
	buf := m.Header.pack()
	buf.Write(encodeUint16(m.messageID))

	_, err := buf.WriteTo(w)
	return err
}

func (m *Unsuback) Unpack(r io.Reader) (err error) {
	m.messageID, err = readUint16(r)
	return
}

func (m Unsuback) String() string {
	return fmt.Sprintf("UNSUBACK(MessageID=%d)", m.messageID)
}
