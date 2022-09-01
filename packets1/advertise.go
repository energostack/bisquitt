package packets1

import (
	"encoding/binary"
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

func (p *Advertise) Unpack(buf []byte) error {
	if len(buf) != int(advertiseVarPartLength) {
		return fmt.Errorf("bad ADVERTISE packet length: expected %d, got %d",
			advertiseVarPartLength, len(buf))
	}

	p.GatewayID = buf[0]
	p.Duration = binary.BigEndian.Uint16(buf[1:3])

	return nil
}

func (p Advertise) String() string {
	return fmt.Sprintf("ADVERTISE(GatewayID=%d,Duration=%d)",
		p.GatewayID, p.Duration)
}
