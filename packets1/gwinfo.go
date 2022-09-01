package packets1

import (
	"fmt"
	"io"

	pkts "github.com/energomonitor/bisquitt/packets"
)

const gwInfoHeaderLength uint16 = 1

type GwInfo struct {
	pkts.Header
	GatewayID      uint8
	GatewayAddress []byte
}

// NOTE: Packet length is initialized in this constructor and recomputed in m.Write().
func NewGwInfo(gatewayID uint8, gatewayAddress []byte) *GwInfo {
	p := &GwInfo{
		Header:         *pkts.NewHeader(pkts.GWINFO, 0),
		GatewayID:      gatewayID,
		GatewayAddress: gatewayAddress,
	}
	p.computeLength()
	return p
}

func (p *GwInfo) computeLength() {
	addrLength := uint16(len(p.GatewayAddress))
	p.Header.SetVarPartLength(gwInfoHeaderLength + addrLength)
}

func (p *GwInfo) Write(w io.Writer) error {
	p.computeLength()

	buf := p.Header.Pack()
	buf.WriteByte(p.GatewayID)
	buf.Write(p.GatewayAddress)

	_, err := buf.WriteTo(w)
	return err
}

func (p *GwInfo) Unpack(buf []byte) error {
	if len(buf) < int(gwInfoHeaderLength) {
		return fmt.Errorf("bad GWINFO packet length: expected >=%d, got %d",
			gwInfoHeaderLength, len(buf))
	}

	p.GatewayID = buf[0]
	p.GatewayAddress = buf[1:]

	return nil
}

func (p GwInfo) String() string {
	return fmt.Sprintf("GWINFO(GatewayID=%d,GatewayAddress=%#v)",
		p.GatewayID, string(p.GatewayAddress))
}
