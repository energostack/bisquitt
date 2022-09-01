package packets1

import (
	"fmt"
	"io"

	pkts "github.com/energomonitor/bisquitt/packets"
)

const advertiseVarPartLength uint16 = 3

type Advertise struct {
	pkts.Header
	GatewayID uint8
	Duration  uint16
}

func NewAdvertise(gatewayID uint8, duration uint16) *Advertise {
	return &Advertise{
		Header:    *pkts.NewHeader(pkts.ADVERTISE, advertiseVarPartLength),
		GatewayID: gatewayID,
		Duration:  duration,
	}
}

func (p *Advertise) Write(w io.Writer) error {
	buf := p.Header.Pack()
	buf.WriteByte(p.GatewayID)
	buf.Write(pkts.EncodeUint16(p.Duration))

	_, err := buf.WriteTo(w)
	return err
}

func (p *Advertise) Unpack(r io.Reader) (err error) {
	if p.GatewayID, err = pkts.ReadByte(r); err != nil {
		return
	}

	p.Duration, err = pkts.ReadUint16(r)
	return
}

func (p Advertise) String() string {
	return fmt.Sprintf("ADVERTISE(GatewayID=%d,Duration=%d)",
		p.GatewayID, p.Duration)
}
