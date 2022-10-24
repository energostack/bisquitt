package packets2

import (
	"fmt"

	pkts "github.com/energomonitor/bisquitt/packets"
)

const searchGwVarPartLength uint16 = 1

type SearchGw struct {
	pkts.Header
	PacketV2
	// Fields
	Radius uint8
}

func NewSearchGw(radius uint8) *SearchGw {
	return &SearchGw{
		Header: *pkts.NewHeader(pkts.SEARCHGW, searchGwVarPartLength),
		Radius: radius,
	}
}

func (p *SearchGw) Pack() ([]byte, error) {
	buf := p.Header.PackToBuffer()

	_ = buf.WriteByte(p.Radius)

	return buf.Bytes(), nil
}

func (p *SearchGw) Unpack(buf []byte) error {
	if len(buf) != int(searchGwVarPartLength) {
		return fmt.Errorf("bad SEARCHGW2 packet length: expected %d, got %d",
			searchGwVarPartLength, len(buf))
	}

	p.Radius = buf[0]

	return nil
}

func (p SearchGw) String() string {
	return fmt.Sprintf("SEARCHGW2(Radius=%d)", p.Radius)
}
