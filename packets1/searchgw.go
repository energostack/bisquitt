package packets1

import (
	"fmt"
	"io"
)

const searchGwVarPartLength uint16 = 1

type SearchGw struct {
	Header
	Radius uint8
}

func NewSearchGw(radius uint8) *SearchGw {
	return &SearchGw{
		Header: *NewHeader(SEARCHGW, searchGwVarPartLength),
		Radius: radius,
	}
}

func (m *SearchGw) Write(w io.Writer) error {
	buf := m.Header.pack()
	buf.WriteByte(m.Radius)

	_, err := buf.WriteTo(w)
	return err
}

func (m *SearchGw) Unpack(r io.Reader) (err error) {
	m.Radius, err = readByte(r)
	return
}

func (m SearchGw) String() string {
	return fmt.Sprintf("SEARCHGW(Radius=%d)", m.Radius)
}
