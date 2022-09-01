package packets1

import (
	"fmt"
	"io"

	pkts "github.com/energomonitor/bisquitt/packets"
)

const connackVarPartLength uint16 = 1

type Connack struct {
	pkts.Header
	ReturnCode ReturnCode
}

func NewConnack(returnCode ReturnCode) *Connack {
	return &Connack{
		Header:     *pkts.NewHeader(pkts.CONNACK, connackVarPartLength),
		ReturnCode: returnCode,
	}
}

func (p *Connack) Write(w io.Writer) error {
	buf := p.Header.Pack()
	buf.WriteByte(byte(p.ReturnCode))

	_, err := buf.WriteTo(w)
	return err
}

func (p *Connack) Unpack(r io.Reader) (err error) {
	var returnCodeByte uint8
	returnCodeByte, err = pkts.ReadByte(r)
	p.ReturnCode = ReturnCode(returnCodeByte)
	return
}

func (p Connack) String() string {
	return fmt.Sprintf("CONNACK(ReturnCode=%d)", p.ReturnCode)
}
