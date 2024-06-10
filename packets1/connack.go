package packets1

import (
	"fmt"

	pkts "github.com/energostack/bisquitt/packets"
)

const connackVarPartLength uint16 = 1

type Connack struct {
	pkts.Header
	// Fields
	ReturnCode ReturnCode
}

func NewConnack(returnCode ReturnCode) *Connack {
	return &Connack{
		Header:     *pkts.NewHeader(pkts.CONNACK, connackVarPartLength),
		ReturnCode: returnCode,
	}
}

func (p *Connack) Pack() ([]byte, error) {
	buf := p.Header.PackToBuffer()

	_ = buf.WriteByte(byte(p.ReturnCode))

	return buf.Bytes(), nil
}

func (p *Connack) Unpack(buf []byte) error {
	if len(buf) != int(connackVarPartLength) {
		return fmt.Errorf("bad CONNACK packet length: expected %d, got %d",
			connackVarPartLength, len(buf))
	}

	p.ReturnCode = ReturnCode(buf[0])

	return nil
}

func (p Connack) String() string {
	return fmt.Sprintf("CONNACK(ReturnCode=%d)", p.ReturnCode)
}
