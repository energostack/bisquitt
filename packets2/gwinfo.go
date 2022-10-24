package packets2

import (
	"fmt"

	pkts "github.com/energomonitor/bisquitt/packets"
)

const gwInfoHeaderLength uint16 = 1

type GwInfo struct {
	pkts.Header
	PacketV2
	// Fields
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

func (p *GwInfo) Pack() ([]byte, error) {
	p.computeLength()
	buf := p.Header.PackToBuffer()

	_ = buf.WriteByte(p.GatewayID)
	_, _ = buf.Write(p.GatewayAddress)

	return buf.Bytes(), nil
}

func (p *GwInfo) Unpack(buf []byte) error {
	if len(buf) < int(gwInfoHeaderLength) {
		return fmt.Errorf("bad GWINFO2 packet length: expected >=%d, got %d",
			gwInfoHeaderLength, len(buf))
	}

	p.GatewayID = buf[0]
	p.GatewayAddress = buf[1:]

	return nil
}

func (p GwInfo) String() string {
	return fmt.Sprintf("GWINFO2(GatewayID=%d,GatewayAddress=%#v)",
		p.GatewayID, string(p.GatewayAddress))
}
