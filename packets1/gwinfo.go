package packets1

import (
	"fmt"
	"io"
)

const gwInfoHeaderLength uint16 = 1

type GwInfo struct {
	Header
	GatewayID      uint8
	GatewayAddress []byte
}

// NOTE: Packet length is initialized in this constructor and recomputed in m.Write().
func NewGwInfo(gatewayID uint8, gatewayAddress []byte) *GwInfo {
	p := &GwInfo{
		Header:         *NewHeader(GWINFO, 0),
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

	buf := p.Header.pack()
	buf.WriteByte(p.GatewayID)
	buf.Write(p.GatewayAddress)

	_, err := buf.WriteTo(w)
	return err
}

func (p *GwInfo) Unpack(r io.Reader) (err error) {
	if p.GatewayID, err = readByte(r); err != nil {
		return
	}

	if p.VarPartLength() > gwInfoHeaderLength {
		p.GatewayAddress = make([]byte, p.VarPartLength()-gwInfoHeaderLength)
		_, err = io.ReadFull(r, p.GatewayAddress)
	} else {
		p.GatewayAddress = nil
	}
	return
}

func (p GwInfo) String() string {
	return fmt.Sprintf("GWINFO(GatewayID=%d,GatewayAddress=%#v)",
		p.GatewayID, string(p.GatewayAddress))
}
