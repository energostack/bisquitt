package packets1

import (
	"io"
)

const pingrespVarPartLength uint16 = 0

type Pingresp struct {
	Header
}

func NewPingresp() *Pingresp {
	return &Pingresp{
		Header: *NewHeader(PINGRESP, pingrespVarPartLength),
	}
}

func (p *Pingresp) Write(w io.Writer) error {
	buf := p.Header.pack()

	_, err := buf.WriteTo(w)
	return err
}

func (p *Pingresp) Unpack(r io.Reader) error {
	return nil
}

func (p Pingresp) String() string {
	return "PINGRESP"
}
