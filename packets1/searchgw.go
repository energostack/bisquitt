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

func (p *SearchGw) Write(w io.Writer) error {
	buf := p.Header.pack()
	buf.WriteByte(p.Radius)

	_, err := buf.WriteTo(w)
	return err
}

func (p *SearchGw) Unpack(r io.Reader) (err error) {
	p.Radius, err = readByte(r)
	return
}

func (p SearchGw) String() string {
	return fmt.Sprintf("SEARCHGW(Radius=%d)", p.Radius)
}
