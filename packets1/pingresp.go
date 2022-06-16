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

func (m *Pingresp) Write(w io.Writer) error {
	buf := m.Header.pack()

	_, err := buf.WriteTo(w)
	return err
}

func (m *Pingresp) Unpack(r io.Reader) error {
	return nil
}

func (m Pingresp) String() string {
	return "PINGRESP"
}
