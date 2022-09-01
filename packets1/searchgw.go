package packets1

import (
	"fmt"
	"io"

	pkts "github.com/energomonitor/bisquitt/packets"
)

const searchGwVarPartLength uint16 = 1

type SearchGw struct {
	pkts.Header
	Radius uint8
}

func NewSearchGw(radius uint8) *SearchGw {
	return &SearchGw{
		Header: *pkts.NewHeader(pkts.SEARCHGW, searchGwVarPartLength),
		Radius: radius,
	}
}

func (p *SearchGw) Write(w io.Writer) error {
	buf := p.Header.Pack()
	buf.WriteByte(p.Radius)

	_, err := buf.WriteTo(w)
	return err
}

func (p *SearchGw) Unpack(buf []byte) error {
	if len(buf) != int(searchGwVarPartLength) {
		return fmt.Errorf("bad SEARCHGW packet length: expected %d, got %d",
			searchGwVarPartLength, len(buf))
	}

	p.Radius = buf[0]

	return nil
}

func (p SearchGw) String() string {
	return fmt.Sprintf("SEARCHGW(Radius=%d)", p.Radius)
}
