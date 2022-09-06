package packets1

import (
	"encoding/binary"
	"fmt"

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

func (p *Advertise) Pack() ([]byte, error) {
	buf := p.Header.PackToBuffer()

	_ = buf.WriteByte(p.GatewayID)
	_, _ = buf.Write(pkts.EncodeUint16(p.Duration))

	return buf.Bytes(), nil
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
