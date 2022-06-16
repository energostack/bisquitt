package messages

import (
	"fmt"
	"io"
)

const searchGwVarPartLength uint16 = 1

type SearchGwMessage struct {
	Header
	Radius uint8
}

func NewSearchGwMessage(radius uint8) *SearchGwMessage {
	return &SearchGwMessage{
		Header: *NewHeader(SEARCHGW, searchGwVarPartLength),
		Radius: radius,
	}
}

func (m *SearchGwMessage) Write(w io.Writer) error {
	buf := m.Header.pack()
	buf.WriteByte(m.Radius)

	_, err := buf.WriteTo(w)
	return err
}

func (m *SearchGwMessage) Unpack(r io.Reader) (err error) {
	m.Radius, err = readByte(r)
	return
}

func (m SearchGwMessage) String() string {
	return fmt.Sprintf("SEARCHGW(Radius=%d)", m.Radius)
}
