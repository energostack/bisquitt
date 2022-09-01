package packets1

import (
	"io"

	pkts "github.com/energomonitor/bisquitt/packets"
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

func (p *Pingresp) Write(w io.Writer) error {
	buf := p.Header.Pack()

	_, err := buf.WriteTo(w)
	return err
}

func (p *Pingresp) Unpack(r io.Reader) error {
	return nil
}

func (p Pingresp) String() string {
	return "PINGRESP"
}
