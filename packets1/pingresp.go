package packets1

import (
	pkts "github.com/energostack/bisquitt/packets"
)

const pingrespVarPartLength uint16 = 0

type Pingresp struct {
	pkts.Header
}

func NewPingresp() *Pingresp {
	return &Pingresp{
		Header: *pkts.NewHeader(pkts.PINGRESP, pingrespVarPartLength),
	}
}

func (p *Pingresp) Pack() ([]byte, error) {
	buf := p.Header.PackToBuffer()
	return buf.Bytes(), nil
}

func (p *Pingresp) Unpack(buf []byte) error {
	return nil
}

func (p Pingresp) String() string {
	return "PINGRESP"
}
